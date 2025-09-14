<script lang="ts">
  import HeaderBar from '$lib/components/HeaderBar.svelte';
</script>

<HeaderBar title="Confidentialité" back={true} />

<main class="max-w-3xl mx-auto px-4 py-8 space-y-6">
  <h1 class="text-2xl font-bold text-gray-900">Politique de confidentialité</h1>

  <p class="text-sm text-gray-700">
    Cette page décrit quelles données sont traitées par l’application Scrabble, dans quel but,
    pour quelle durée, et vos droits sur ces données. Elle couvre le site web (frontend) et l’API backend.
  </p>

  <section class="space-y-3">
    <h2 class="text-xl font-semibold">Données traitées</h2>
    <ul class="list-disc pl-6 text-sm text-gray-800 space-y-1">
      <li>Compte utilisateur : nom d’utilisateur (username), mot de passe (haché côté serveur), rôle (ex. admin).</li>
      <li>Jeux : parties créées/rejointes, état de la grille, rack, scores, historique des coups, statut de partie.</li>
      <li>Tickets (reports) : titre, contenu, type, statut, auteur et horodatage.</li>
      <li>Notifications Push (optionnel) : abonnement Web Push (VAPID) lié à votre compte/appareil.</li>
      <li>Journaux techniques (logs) : événements techniques pour diagnostic/qualité de service.</li>
    </ul>
  </section>

  <section class="space-y-3">
    <h2 class="text-xl font-semibold">Traitements côté navigateur</h2>
    <ul class="list-disc pl-6 text-sm text-gray-800 space-y-1">
      <li>Session locale : stockage du token JWT et du nom d’utilisateur dans le localStorage pour rester connecté.</li>
      <li>Service Worker : mise en cache d’actifs pour de meilleures performances et page hors‑ligne.</li>
      <li>Notifications : si vous consentez, l’app enregistre un abonnement Push sur cet appareil.</li>
      <li>Cookies : l’app n’utilise pas de cookies de suivi/analytics.</li>
    </ul>
  </section>

  <section class="space-y-3">
    <h2 class="text-xl font-semibold">Finalités et bases légales</h2>
    <ul class="list-disc pl-6 text-sm text-gray-800 space-y-1">
      <li>Fournir le service (création/gestion de parties, historique, tickets) : exécution du service demandé.</li>
      <li>Notifications Push (tour/messagerie, si activées) : consentement.</li>
      <li>Sécurité/qualité (logs, anti‑abus) : intérêt légitime.</li>
    </ul>
  </section>

  <section class="space-y-3">
    <h2 class="text-xl font-semibold">Durées de conservation</h2>
    <ul class="list-disc pl-6 text-sm text-gray-800 space-y-1">
      <li>Compte : jusqu’à suppression du compte (ou demande explicite).</li>
      <li>Parties et historique : nécessaires au fonctionnement affichage/classements, puis archivage/suppression périodique.</li>
      <li>Tickets : pendant le traitement puis archivage raisonnable.</li>
      <li>Abonnements Push : jusqu’à désinscription ou invalidation par le navigateur.</li>
      <li>Logs techniques : durée limitée (courte) strictement pour diagnostic.</li>
    </ul>
  </section>

  <section class="space-y-3">
    <h2 class="text-xl font-semibold">Partage des données</h2>
    <p class="text-sm text-gray-700">
      Les données ne sont pas revendues. Elles peuvent transiter via :
    </p>
    <ul class="list-disc pl-6 text-sm text-gray-800 space-y-1">
      <li>Prestataire d’hébergement (base de données/API/serveur web).</li>
      <li>Services Push des navigateurs (infrastructure VAPID/Push gérée par le fournisseur du navigateur).</li>
    </ul>
  </section>

  <section class="space-y-3">
    <h2 class="text-xl font-semibold">Vos droits</h2>
    <ul class="list-disc pl-6 text-sm text-gray-800 space-y-1">
      <li>Accès, rectification, suppression de votre compte et de vos tickets.</li>
      <li>Retrait du consentement aux notifications (désactivez dans votre profil, et/ou via les réglages du navigateur).</li>
      <li>Opposition et limitation (dans les limites techniques et légales).</li>
    </ul>
    <p class="text-sm text-gray-700">
      Pour exercer vos droits (ou demander la suppression du compte), contactez :
      <a class="text-green-700 underline" href="mailto:leroyerbaptiste@gmail.com">leroyerbaptiste@gmail.com</a>.
    </p>
  </section>

  <section class="space-y-3">
    <h2 class="text-xl font-semibold">Sécurité</h2>
    <ul class="list-disc pl-6 text-sm text-gray-800 space-y-1">
      <li>Mots de passe hachés côté serveur, tokens JWT côté client.</li>
      <li>Accès API via en‑tête Authorization ; séparation Front/API.</li>
      <li>Bonnes pratiques courantes (contrôles d’accès, validations, journalisation).</li>
    </ul>
  </section>

  <section class="space-y-1">
    <h2 class="text-xl font-semibold">Contact</h2>
    <p class="text-sm text-gray-700">
      Email : <a class="text-green-700 underline" href="mailto:leroyerbaptiste@gmail.com">leroyerbaptiste@gmail.com</a>
    </p>
  </section>

  <p class="text-xs text-gray-500">
    Dernière mise à jour : {new Date().toLocaleDateString('fr-FR')}
  </p>
</main>