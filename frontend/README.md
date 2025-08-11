# Scrabble — Frontend (SvelteKit + TS)

Interface web du jeu de Scrabble en ligne. Cette app consomme l’API, gère l’auth (JWT côté client), l’affichage du plateau, la création/gestion de parties, l’historique, la création de tickets, et les notifications push via un **Service Worker**. PWA-ready (manifest + bannière d’installation), UI Tailwind.

---

## Fonctionnalités

* **Auth** : login/register, persistance locale (`localStorage`) et auto‑redirection vers `/login` si 401.
* **Parties** : création, liste filtrée (à mon tour / en cours / terminées), renommage, suppression.
* **Plateau 15×15** : cases spéciales (DL, TL, DW, TW, ★), aperçu du **score du coup** en direct, surlignage du dernier coup, drag‑and‑drop pour réordonner son rack.
* **Historique** : liste chronologique des coups (mot, direction, score, positions), accès depuis la partie.
* **Reports** : création et consultation de ses tickets (statut, type, date).
* **PWA** : manifest, bannière d’installation (A2HS), thème, icônes.
* **Notifications Push** : abonnement VAPID + réception via Service Worker.

---

## Stack & Organisation

* **SvelteKit** (TypeScript)
* **TailwindCSS** (styles utilitaires via `src/app.css`)
* **Axios** pour les appels REST (`$lib/api`)
* **svelte-dnd-action** pour le DnD du rack

Arborescence (extrait) :

```
frontend/
  src/
    lib/
      api.ts              # client Axios + intercepteur 401
      cells.ts            # cases spéciales (DL/TL/DW/TW/★)
      lettres_value.ts    # valeurs des lettres FR + computeWordValue
      push.ts             # abonnement VAPID
      components/
        Board.svelte      # plateau 15×15 + surlignage dernier coup
        GameList.svelte   # liste de parties
        GameItem.svelte   # item de partie
      stores/
        user.ts           # store utilisateur (localStorage)
        pendingMove.ts    # lettres posées en attente
        game.ts           # cache de la partie courante
      types/
        game_infos.ts     # types de GameInfo (détails partie)
        game_summary.ts   # types de GameSummary (liste)
    routes/
      +layout.svelte      # Navbar, SEO/OG meta
      +page.svelte        # accueil (loginé vs non logué) + bannière PWA
      login/+page.svelte  # connexion
      register/+page.svelte
      impersonate/+page.svelte  # utilitaire d’admin/dev
      admin/change-password/+page.svelte
      games/new/+page.svelte    # création
      games/[id]/+page.svelte   # vue partie (plateau, rack, actions)
      games/[id]/history/+page.svelte # historique
      report/+page.svelte        # mes tickets
      report/new/+page.svelte    # créer un ticket
      report/[id]/+page.svelte   # détail d’un ticket
  static/
    manifest.json
    icons/icon-192.png
    icons/icon-512.png
  service-worker.js
  Dockerfile
```

---

## Pré‑requis

* Node.js 20+ (recommandé) et npm
* API démarrée et accessible (voir README backend)

---

## Configuration (env)

Créer `frontend/.env` (exemple) :

```env
PUBLIC_API_BASE_URL="http://localhost:8888/"
PUBLIC_VAPID_PUBLIC_KEY="<clé publique VAPID>"
```

> **Important** : SvelteKit expose au client uniquement les variables préfixées par `PUBLIC_`. Le code utilise `PUBLIC_API_BASE_URL` et `PUBLIC_VAPID_PUBLIC_KEY`.

> Si vous utilisez `docker-compose` (service `front`), harmonisez l’ENV : préférez `PUBLIC_API_BASE_URL` (remplacez les références à `VITE_API_BASE_URL`).

---

## Démarrer en développement

Depuis la racine du repo :

```bash
make front  # équivaut à: cd frontend && npm run dev
```

ou directement :

```bash
cd frontend
npm install
npm run dev
```

L’app écoute par défaut sur `http://localhost:5173/` (Vite dev server).

### Auth

* L’utilisateur logué est stocké dans `localStorage` (`{ username, token }`).
* L’intercepteur Axios efface la session et redirige vers `/login` si la réponse = 401.

---

## Build & Production

Compilation :

```bash
cd frontend
npm run build
npm run preview  # test local
```

### Docker (adapter‑node)

Le `Dockerfile` produit une image Node qui lance `node build` sur le port **3000** :

```bash
docker build -t scrabble-frontend ./frontend
docker run -p 3000:3000 --env-file ./frontend/.env scrabble-frontend
```

> Configurez `PUBLIC_API_BASE_URL` et `PUBLIC_VAPID_PUBLIC_KEY` dans l’environnement du conteneur.

---

## Service Worker & Notifications Push

* **Service Worker** : `src/service-worker.js` gère l’activation et l’événement `push` (affichage de notifications) et `notificationclick` (navigation).
* **Abonnement** : `src/lib/push.ts` demande la permission, récupère `navigator.serviceWorker.ready`, puis s’abonne via `PushManager.subscribe()` avec la **clé VAPID publique** (`PUBLIC_VAPID_PUBLIC_KEY`). L’abonnement est envoyé à l’API (`/notifications/push-subscribe`).
* **Bannière PWA** : l’accueil (`+page.svelte`) affiche une bannière « Installer » et propose d’activer les notifications.

> Si les notifications ne s’affichent pas, vérifiez que le Service Worker est bien **enregistré** sur votre domaine (vous pouvez ajouter l’enregistrement dans `+layout.svelte` ou `hooks.client.ts` : `navigator.serviceWorker.register('/service-worker.js')`).

---

## Pages & UX

* **Accueil** : selon la session, affiche soit `NoLoginPage`, soit `LoginedPage`.
* **LoginedPage** :

  * filtre les parties (à mon tour / en cours / terminées),
  * actions : **Créer**, **Renommer**, **Supprimer**.
* **Vue Partie** (`/games/[id]`) :

  * **Plateau** (Board) :

    * affiche lettres posées (anciens vs dernier coup surligné),
    * clique pour poser/retirer une lettre en attente,
    * cases spéciales & points visibles.
  * **Rack** : drag‑and‑drop fluide pour réordonner, sélection de lettre.
  * **Actions** : Annuler, Passer, Échanger, **Valider** (envoi à l’API), bouton **Classement** (modal) + lien **Historique**.
* **Historique** : détail des coups (mot, direction H/V, score, date/heure) et positions des lettres.
* **Reports** : création d’un ticket (bug/suggestion/feedback/other) et liste « mes tickets ».
* **Admin** : page de changement de mot de passe.

---

## Client API (Axios)

* `baseURL` : `PUBLIC_API_BASE_URL`.
* **Auth Header** : injecté automatiquement si utilisateur présent.
* **Intercepteur 401** : clear du store + redirection `/login`.

Endpoints consommés (exemples) :

* Auth : `POST /auth/login`, `POST /auth/register`, `POST /auth/change-password`, `GET /auth/connect-as` (dev/impersonate).
* Game : `GET /game`, `POST /game`, `GET /game/:id`, `POST /game/:id/play`, `POST /game/:id/pass`, `GET /game/:id/new_rack`, `POST /game/:id/simulate_score`, `PUT /game/:id/rename`, `DELETE /game/:id`.
* Users : `GET /users/suggest?q=`.
* Reports : `GET /report/me`, `GET /report/:id`, `POST /report`, `PATCH /report/:id`.
* Notifications : `POST /notifications/push-subscribe`.

---

## Conseils & Bonnes pratiques

* **Variables d’environnement** : n’utilisez que des `PUBLIC_*` côté client. Évitez les `VITE_*` si vous accédez via `$env/dynamic/public`.
* **CORS** : si API sur un autre domaine/port, autoriser l’origine front dans l’API.
* **Accessibilité** :

  * tailles/contrastes des tuiles et points visibles,
  * boutons et liens avec libellés clairs,
  * transitions non bloquantes (DnD → `flip`).
* **Resilience** : les actions critiques (valider, échanger, passer) gèrent les erreurs réseau via alert/toast.

---

## Dépannage (FAQ)

**Je suis redirigé vers /login** → le token a expiré ou est absent : reconnectez‑vous.

**Le score affiché reste à 0** → vérifiez que des lettres sont posées et que l’API `/simulate_score` répond (voir logs réseau).

**Les notifications ne marchent pas** →

* assurez la présence/registration du Service Worker,
* vérifiez `PUBLIC_VAPID_PUBLIC_KEY`,
* confirmez l’abonnement côté API (endpoint `/notifications/push-subscribe`).

**Impossible d’inviter un joueur** → la suggestion apparaît après 2 caractères, l’API `/users/suggest` doit être accessible.

---

## Makefile (racine)

* `make front` : lance le frontend en mode dev.
* `make air` : API en live reload.
* `make db` : base Postgres via Docker Compose.
* `make migrate-up|down` : migrations DB.

---

## Licence

Projet éducatif/démo — choisissez la licence adaptée à votre cas.
