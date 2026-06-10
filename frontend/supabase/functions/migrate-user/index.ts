import { serve } from "https://deno.land/std@0.168.0/http/server.ts";
import { createClient } from "https://esm.sh/@supabase/supabase-js@2";
import bcrypt from "https://esm.sh/bcryptjs@2.4.3";

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers": "authorization, x-client-info, apikey, content-type",
};

serve(async (req) => {
  // Handle CORS preflight
  if (req.method === "OPTIONS") {
    return new Response("ok", { headers: corsHeaders });
  }

  try {
    const supabaseUrl = Deno.env.get("SUPABASE_URL") ?? "";
    const supabaseServiceKey = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY") ?? "";
    const supabase = createClient(supabaseUrl, supabaseServiceKey);

    const body = await req.json();
    const action = body.action || "login";
    const username = body.username;

    if (!username) {
      return new Response(
        JSON.stringify({ error: "Username is required" }),
        { status: 400, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    const cleanUsername = username.trim();
    const email = `${cleanUsername.toLowerCase()}@scrabble-game.com`;

    if (action === "connect-as") {
      // Impersonation: Admin connects as another user
      const authHeader = req.headers.get("Authorization");
      if (!authHeader) {
        return new Response(
          JSON.stringify({ error: "Missing authorization header" }),
          { status: 401, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      const token = authHeader.replace("Bearer ", "");
      const { data: { user: callerUser }, error: callerError } = await supabase.auth.getUser(token);
      if (callerError || !callerUser) {
        return new Response(
          JSON.stringify({ error: "Unauthorized caller" }),
          { status: 401, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      // Check if caller is admin
      const { data: callerProfile, error: callerProfileError } = await supabase
        .from("users")
        .select("role")
        .eq("uuid", callerUser.id)
        .single();

      if (callerProfileError || callerProfile?.role !== "admin") {
        return new Response(
          JSON.stringify({ error: "Forbidden: Admins only" }),
          { status: 403, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      // Ensure the target user exists (and migrate if they exist only in old db)
      const { data: targetProfile, error: targetError } = await supabase
        .from("users")
        .select("id, uuid, password")
        .eq("username", cleanUsername)
        .maybeSingle();

      if (targetError) {
        return new Response(
          JSON.stringify({ error: `Target lookup error: ${targetError.message}` }),
          { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      if (!targetProfile) {
        return new Response(
          JSON.stringify({ error: "Target user not found" }),
          { status: 404, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      // If they exist in users but uuid is null (not yet migrated to Supabase Auth)
      if (!targetProfile.uuid) {
        // We migrate them by signing them up with a random password since we are admin
        const randomPassword = crypto.randomUUID();
        const { error: signUpError } = await supabase.auth.signUp({
          email,
          password: randomPassword,
          options: {
            data: {
              username: cleanUsername,
            },
          },
        });

        if (signUpError) {
          return new Response(
            JSON.stringify({ error: `Migration signup for impersonation failed: ${signUpError.message}` }),
            { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
          );
        }
      }

      // Generate magiclink link & OTP
      const { data: linkData, error: linkError } = await supabase.auth.admin.generateLink({
        type: "magiclink",
        email,
      });

      if (linkError || !linkData?.properties?.email_otp) {
        return new Response(
          JSON.stringify({ error: `Failed to generate login link: ${linkError?.message || "No OTP returned"}` }),
          { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      // Verify OTP to get session access token
      const { data: sessionData, error: sessionError } = await supabase.auth.verifyOtp({
        email,
        token: linkData.properties.email_otp,
        type: "magiclink",
      });

      if (sessionError || !sessionData?.session) {
        return new Response(
          JSON.stringify({ error: `Failed to verify login link: ${sessionError?.message || "No session"}` }),
          { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      return new Response(
        JSON.stringify({
          token: sessionData.session.access_token,
          session: sessionData.session,
          user: sessionData.user,
        }),
        { status: 200, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }
    if (action === "register") {
      const password = body.password;
      if (!password) {
        return new Response(
          JSON.stringify({ error: "Password is required" }),
          { status: 400, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      // Check if username already exists in public.users
      const { data: existingUser, error: checkError } = await supabase
        .from("users")
        .select("id")
        .eq("username", cleanUsername)
        .maybeSingle();

      if (checkError) {
        return new Response(
          JSON.stringify({ error: `Database error: ${checkError.message}` }),
          { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      if (existingUser) {
        return new Response(
          JSON.stringify({ error: "Username is already taken" }),
          { status: 400, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      // Create new user using admin client (auto-confirm email)
      const { data: adminUserData, error: authError } = await supabase.auth.admin.createUser({
        email,
        password,
        email_confirm: true,
        user_metadata: {
          username: cleanUsername,
        },
      });

      if (authError) {
        return new Response(
          JSON.stringify({ error: `Registration failed: ${authError.message}` }),
          { status: 400, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      // Sign in to get session
      const { data: signInData, error: signInError } = await supabase.auth.signInWithPassword({
        email,
        password,
      });

      if (signInError) {
        return new Response(
          JSON.stringify({ error: `Registration signin failed: ${signInError.message}` }),
          { status: 400, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      return new Response(
        JSON.stringify({ session: signInData.session, user: signInData.user, registered: true }),
        { status: 200, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    // Default action: login/migration
    const password = body.password;
    if (!password) {
      return new Response(
        JSON.stringify({ error: "Password is required" }),
        { status: 400, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    // 1. Check if the user exists in public.users
    const { data: dbUser, error: dbError } = await supabase
      .from("users")
      .select("id, password, uuid")
      .eq("username", cleanUsername)
      .maybeSingle();

    if (dbError) {
      return new Response(
        JSON.stringify({ error: `Database error: ${dbError.message}` }),
        { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    // Case A: User exists and has NOT been migrated yet (uuid is null)
    if (dbUser && !dbUser.uuid) {
      // Verify bcrypt password
      const passwordMatch = bcrypt.compareSync(password, dbUser.password);
      if (!passwordMatch) {
        return new Response(
          JSON.stringify({ error: "Invalid credentials" }),
          { status: 401, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      // Migrate: Create auth user in Supabase Auth using admin client (auto-confirm email)
      const { data: adminUserData, error: authError } = await supabase.auth.admin.createUser({
        email,
        password,
        email_confirm: true,
        user_metadata: {
          username: cleanUsername,
        },
      });

      if (authError) {
        return new Response(
          JSON.stringify({ error: `Migration signup failed: ${authError.message}` }),
          { status: 400, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      // Sign in to get session
      const { data: signInData, error: signInError } = await supabase.auth.signInWithPassword({
        email,
        password,
      });

      if (signInError) {
        return new Response(
          JSON.stringify({ error: `Migration signin failed: ${signInError.message}` }),
          { status: 400, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      return new Response(
        JSON.stringify({ session: signInData.session, user: signInData.user, migrated: true }),
        { status: 200, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    // Case B: User already migrated or is a new user
    // We just attempt standard sign-in via Supabase Auth
    let { data: signInData, error: signInError } = await supabase.auth.signInWithPassword({
      email,
      password,
    });

    if (signInError && signInError.message === "Email not confirmed") {
      // Fallback: Auto-confirm email using admin API and try again
      if (dbUser && dbUser.uuid) {
        const { error: confirmError } = await supabase.auth.admin.updateUserById(dbUser.uuid, {
          email_confirm: true,
        });
        if (!confirmError) {
          // Retry sign-in
          const retryResult = await supabase.auth.signInWithPassword({
            email,
            password,
          });
          signInData = retryResult.data;
          signInError = retryResult.error;
        }
      }
    }

    if (signInError) {
      return new Response(
        JSON.stringify({ error: signInError.message }),
        { status: 401, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    return new Response(
      JSON.stringify({ session: signInData.session, user: signInData.user, migrated: false }),
      { status: 200, headers: { ...corsHeaders, "Content-Type": "application/json" } }
    );

  } catch (err) {
    return new Response(
      JSON.stringify({ error: err.message }),
      { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
    );
  }
});
