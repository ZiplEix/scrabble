/// <reference types="cypress" />

describe('Authentification - Connexion', () => {
    const baseUrl = Cypress.config('baseUrl') || 'http://localhost:3000';

    const visitLogin = () => cy.visit(`${baseUrl}/login`);

    const fillLoginForm = (username: string, password: string) => {
        cy.get('input[placeholder="Nom d\'utilisateur"]').clear().type(username);
        cy.get('input[placeholder="Mot de passe"]').clear().type(password);
    };

    beforeEach(() => {
        cy.clearLocalStorage();
    });

    it('se connecte avec succès et redirige vers l\'accueil', () => {
        // Éviter qu'un GET /game renvoie 401 et purge le store
        cy.intercept('GET', '**/game', {
            statusCode: 200,
            headers: { 'content-type': 'application/json' },
            body: { games: [] },
        }).as('getGames');

        cy.intercept('POST', '**/auth/login', {
            statusCode: 200,
            headers: { 'content-type': 'application/json' },
            body: { token: 'fake-jwt-token' },
        }).as('login');

        visitLogin();
        fillLoginForm('Jean.Dupont', 'Password123!');

        cy.get('button[type="submit"]').click();

        cy.wait('@login');
        cy.wait('@getGames');
        cy.location('pathname').should('eq', '/');

        // le store sauvegarde en localStorage sous la clé "user"
        cy.window().then((win) => {
            const raw = win.localStorage.getItem('user');
            expect(raw, 'item user dans localStorage').to.be.a('string');
            const parsed = raw ? JSON.parse(raw) : null;
            expect(parsed).to.include({ token: 'fake-jwt-token' });
            // le login normalise le username en minuscule/trim
            expect(parsed.username).to.eq('jean.dupont');
        });
    });

    it('affiche une erreur quand les identifiants sont invalides', () => {
        cy.intercept('POST', '**/auth/login', {
            statusCode: 401,
            headers: { 'content-type': 'application/json' },
            body: { message: 'Identifiants invalides' },
        }).as('loginError');

        visitLogin();
        fillLoginForm('baduser', 'wrong');
        cy.get('button[type="submit"]').click();

        cy.wait('@loginError');
        cy.location('pathname').should('eq', '/login');
        cy.contains('p', 'Identifiants invalides').should('be.visible');
    });
});

