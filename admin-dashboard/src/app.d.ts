// See https://svelte.dev/docs/kit/types#app.d.ts

import type { User } from "./routes/dashboard/users/type";

// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}

		interface PageState {
			user?: User;
		}

		interface PageData {
			user: User;
		}
	}
}

export {};
