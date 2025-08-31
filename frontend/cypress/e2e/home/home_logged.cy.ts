/// <reference types="cypress" />

describe('Accueil — utilisateur connecté', () => {
    const baseUrl = Cypress.config('baseUrl') || 'http://localhost:3000';

    const gamesPayload = {
        games: [
            {
                id: '30283156-23d3-476a-a415-43ecc7a58599',
                name: 'Aude coco hort – 3 – revanche',
                status: 'ongoing',
                current_turn_user_id: 2,
                current_turn_username: 'hobulala',
                last_play_time: '2025-08-29T13:11:03.937592Z',
                is_your_game: false,
            },
            {
                id: '93589028-f5e7-4172-a547-82a3cabc62aa',
                name: 'Aude et coco  – revanche',
                status: 'ongoing',
                current_turn_user_id: 2,
                current_turn_username: 'coco',
                last_play_time: '2025-08-29T13:10:04.382512Z',
                is_your_game: true,
            },
            {
                id: 'c7e7ca46-0dff-462d-a26d-4e649ae2c4d3',
                name: 'Aude coco hort – revanche – revanche – revanche',
                status: 'ongoing',
                current_turn_user_id: 3,
                current_turn_username: 'hobulala',
                last_play_time: '2025-08-28T20:38:06.364402Z',
                is_your_game: true,
            },
            {
                id: '91bb2c54-e7a0-4d2e-af62-58ecb597ada0',
                name: 'Hortense et Aude – revanche – revanche',
                status: 'ongoing',
                current_turn_user_id: 3,
                current_turn_username: 'hobulala',
                last_play_time: '2025-08-28T20:14:01.177475Z',
                is_your_game: false,
            },
            {
                id: '04d3c332-9d1e-46d1-b4c4-5325d95ea222',
                name: 'Aude et coco ',
                status: 'ended',
                current_turn_user_id: 2,
                current_turn_username: 'aude',
                last_play_time: '2025-08-28T20:09:35.484652Z',
                is_your_game: true,
                winner_username: 'aude',
            },
            {
                id: 'aa945e6f-2216-4d62-a69b-ef8246ba7131',
                name: 'Hortense et Aude – revanche',
                status: 'ended',
                current_turn_user_id: 2,
                current_turn_username: 'aude',
                last_play_time: '2025-08-26T06:38:53.288673Z',
                is_your_game: false,
                winner_username: 'hobulala',
            },
            {
                id: 'e25f49e3-b4ec-4c42-834b-1f264fca878c',
                name: 'test chat',
                status: 'ongoing',
                current_turn_user_id: 1,
                current_turn_username: 'baptiste',
                last_play_time: '2025-08-25T19:58:39.191131Z',
                is_your_game: true,
            },
            {
                id: 'eebd59a5-2702-46a5-a64f-b262c12991a4',
                name: 'Aude et coco  16/08/25 – revanche',
                status: 'ended',
                current_turn_user_id: 7,
                current_turn_username: 'coco',
                last_play_time: '2025-08-24T04:59:20.121202Z',
                is_your_game: true,
                winner_username: 'coco',
            },
            {
                id: '4a6d63ab-4e73-4f58-8858-f74d6c937130',
                name: 'Aude et coco  16/08/25 – revanche',
                status: 'ended',
                current_turn_user_id: 7,
                current_turn_username: 'coco',
                last_play_time: '2025-08-23T20:27:14.306891Z',
                is_your_game: false,
                winner_username: 'coco',
            },
            {
                id: '5f034c3c-eee8-4c98-a385-f62bd9e163b3',
                name: 'Aude coco hort – 3',
                status: 'ended',
                current_turn_user_id: 7,
                current_turn_username: 'coco',
                last_play_time: '2025-08-22T15:32:05.243445Z',
                is_your_game: false,
                winner_username: 'aude',
            },
            {
                id: '8d43d9d6-270f-4974-9e2b-112038d2b971',
                name: 'Aude coco hort – revanche – revanche',
                status: 'ended',
                current_turn_user_id: 2,
                current_turn_username: 'aude',
                last_play_time: '2025-08-21T16:58:11.46675Z',
                is_your_game: true,
                winner_username: 'coco',
            },
            {
                id: '00aee363-69e4-4afa-ac4f-da6e8d55896b',
                name: 'Aude et coco  16/08/25',
                status: 'ended',
                current_turn_user_id: 7,
                current_turn_username: 'coco',
                last_play_time: '2025-08-19T20:24:07.456609Z',
                is_your_game: true,
                winner_username: 'coco',
            },
            {
                id: '22e9ee27-eca3-45a9-b7ff-4731d1aac307',
                name: 'Test en cours',
                status: 'ongoing',
                current_turn_user_id: 4,
                current_turn_username: 'jerome',
                last_play_time: '2025-08-17T17:00:50.853276Z',
                is_your_game: false,
            },
            {
                id: '913370e1-acee-49d0-b745-8914d018431a',
                name: 'Aude et Jérôme ',
                status: 'ongoing',
                current_turn_user_id: 4,
                current_turn_username: 'jerome',
                last_play_time: '2025-08-17T15:45:57.152309Z',
                is_your_game: true,
            },
            {
                id: '37fd0ce2-4f5b-46e6-aeef-c1cc0bc31ea2',
                name: 'Aude et coco – revanche',
                status: 'ended',
                current_turn_user_id: 2,
                current_turn_username: 'aude',
                last_play_time: '2025-08-16T09:19:24.775866Z',
                is_your_game: false,
                winner_username: 'coco',
            },
            {
                id: 'a27e3d81-b6fe-4be7-8114-2346f3dd30ea',
                name: 'Nous 3',
                status: 'ended',
                current_turn_user_id: 2,
                current_turn_username: 'aude',
                last_play_time: '2025-08-15T14:40:07.231906Z',
                is_your_game: true,
                winner_username: 'hobulala',
            },
            {
                id: '1fed7cd0-2c53-4a98-a19e-1ce516c8e8db',
                name: 'Aude coco hort – revanche',
                status: 'ended',
                current_turn_user_id: 2,
                current_turn_username: 'aude',
                last_play_time: '2025-08-10T20:39:48.524056Z',
                is_your_game: true,
                winner_username: 'aude',
            },
            {
                id: '0c4bf01b-1d47-47c2-b1d0-44f4dfbd4077',
                name: 'Hortense et Aude',
                status: 'ended',
                current_turn_user_id: 3,
                current_turn_username: 'hobulala',
                last_play_time: '2025-08-10T06:32:06.010751Z',
                is_your_game: true,
                winner_username: 'hobulala',
            },
            {
                id: 'fb565ae7-2a7b-4609-9e6a-b3d76d2b994b',
                name: 'Aude et coco',
                status: 'ended',
                current_turn_user_id: 2,
                current_turn_username: 'aude',
                last_play_time: '2025-08-09T12:26:06.099846Z',
                is_your_game: true,
                winner_username: 'aude',
            },
            {
                id: 'da29f576-c445-4764-b99f-046f734dc2e9',
                name: 'Coco et Aude',
                status: 'ended',
                current_turn_user_id: 2,
                current_turn_username: 'aude',
                last_play_time: '2025-08-09T12:25:52.247894Z',
                is_your_game: true,
                winner_username: 'coco',
            },
        ],
    } as const;

    beforeEach(() => {
        cy.clearLocalStorage();
        // Simule un utilisateur déjà connecté
        const user = { username: 'baptiste', token: 'jwt-token' };
        cy.window().then((win) => {
            win.localStorage.setItem('user', JSON.stringify(user));
        });

        cy.intercept('GET', '**/game', {
            statusCode: 200,
            headers: { 'content-type': 'application/json' },
            body: gamesPayload,
        }).as('getGames');
    });

    const closeNewsBannerIfPresent = () => {
        cy.get('body').then(($body) => {
            const closeBtn = $body.find('button[aria-label="Fermer"]');
            if (closeBtn.length) {
                cy.wrap(closeBtn.first()).click({ force: true });
            }
        });
    };

    it('affiche correctement les sections et les parties', () => {
        cy.visit(`${baseUrl}/`);
        cy.wait('@getGames');
        closeNewsBannerIfPresent();

    // Onglet "À mon tour" (par défaut) — la partie "test chat" doit y être avec le bon libellé
        cy.contains('h2', 'test chat')
            .parents('div.relative')
            .first()
            .within(() => {
                cy.contains('Tour de :').should('contain', 'baptiste');
                cy.contains('Dernier coup :').should('exist');
                cy.contains('Gagnant:').should('not.exist');
                cy.get('button[aria-label="Menu"]').should('exist'); // is_your_game: true
            });

    // Onglet "En cours" (pas à mon tour)
    cy.contains('button', 'En cours').click();
    cy.contains('h2', 'Aude coco hort – 3 – revanche')
            .parents('div.relative')
            .first()
            .within(() => {
                cy.contains('Tour de :').should('contain', 'hobulala');
                cy.get('button[aria-label="Menu"]').should('not.exist'); // is_your_game: false
            });

    // Onglet "Terminées" — "Nous 3" doit afficher le gagnant et pas le tour
    cy.contains('button', 'Terminées').click();
        cy.contains('h2', 'Nous 3')
            .parents('div.relative')
            .first()
            .within(() => {
                cy.contains('Gagnant:').should('contain', 'hobulala');
                cy.contains('Tour de :').should('not.exist');
                cy.get('button[aria-label="Menu"]').should('exist'); // is_your_game: true
            });
    });

    it('navigue vers la création de partie', () => {
    cy.visit(`${baseUrl}/`);
    cy.wait('@getGames');
    closeNewsBannerIfPresent();
    cy.get('button[aria-label="Créer une nouvelle partie"]').click();
        cy.location('pathname').should('eq', '/games/new');
    });
});
