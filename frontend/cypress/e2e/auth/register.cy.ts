/// <reference types="cypress" />

describe('Authentification - Inscription', () => {
    const baseUrl = Cypress.config('baseUrl') || 'http://localhost:3000';

    const visitRegister = () => cy.visit(`${baseUrl}/register`);

    const fillRegisterForm = (username: string, password: string) => {
        cy.get('input[placeholder="Nom d\'utilisateur"]').clear().type(username);
        cy.get('input[placeholder="Mot de passe"]').clear().type(password);
    };

    beforeEach(() => {
        cy.clearLocalStorage();
    });

    beforeEach(() => {
        cy.clearLocalStorage();
    });

    it("crée un compte et redirige vers l'accueil", () => {
        // La page d'accueil (LoginedPage) appelle GET /game au mount — on le mocke pour éviter un 401 qui viderait le store
        cy.intercept('GET', '**/game', {
            statusCode: 200,
            headers: { 'content-type': 'application/json' },
            body: { games: [] },
        }).as('getGames');

        cy.intercept('POST', '**/auth/register', (req) => {
            // vérifier que le frontend envoie bien { username: <lowercase>, password }
            expect(req.headers['content-type']).to.include('application/json');
            const body = req.body as { username: string; password: string };
            expect(body.username).to.eq('alice');
            expect(body.password).to.eq('StrongPass!');
            req.reply({
                statusCode: 201,
                headers: { 'content-type': 'application/json' },
                body: { token: 'new-user-jwt' },
            });
        }).as('register');

        visitRegister();
        fillRegisterForm('Alice', 'StrongPass!');
        cy.get('button[type="submit"]').click();

        cy.wait('@register');
        // Attendre l'appel de /game déclenché par l'accueil après redirection
        cy.wait('@getGames');
        cy.location('pathname').should('eq', '/');

        cy.window().then((win) => {
            const raw = win.localStorage.getItem('user');
            expect(raw, 'item user dans localStorage').to.be.a('string');
            const parsed = raw ? JSON.parse(raw) : null;
            // le frontend normalise en minuscule à l'inscription
            expect(parsed).to.include({ token: 'new-user-jwt', username: 'alice' });
        });
    });

    it("affiche un message si le nom d'utilisateur est déjà pris", () => {
        cy.intercept('POST', '**/auth/register', {
            statusCode: 409,
            headers: { 'content-type': 'application/json' },
            body: {
                error:
                    'username alice already exists: failed to execute statement: pq: duplicate key value violates unique constraint "users_username_key"',
                message: "Le nom d'utilisateur existe déjà, veuillez en choisir un autre",
            },
        }).as('registerError');

        visitRegister();
        fillRegisterForm('Alice', 'StrongPass!');
        cy.get('button[type="submit"]').click();

        cy.wait('@registerError').then((interception) => {
            const msg = interception?.response?.body?.message as string;
            expect(msg).to.eq("Le nom d'utilisateur existe déjà, veuillez en choisir un autre");
            // Rechercher le texte où qu'il soit dans le DOM (pas uniquement dans <p>)
            cy.contains(new RegExp(msg.replace(/[-/\\^$*+?.()|[\]{}]/g, '.'), 'i')).should('be.visible');
        });
        cy.location('pathname').should('eq', '/register');
    });

    it("affiche un message si le nom d'utilisateur est vide (espaces)", () => {
        cy.intercept('POST', '**/auth/register', {
            statusCode: 400,
            headers: { 'content-type': 'application/json' },
            body: { message: "Le nom d'utilisateur est requis" },
        }).as('registerEmpty');

        visitRegister();
        fillRegisterForm('   ', 'StrongPass!');
        cy.get('button[type="submit"]').click();

        cy.wait('@registerEmpty').then((interception) => {
            const msg = interception?.response?.body?.message as string;
            expect(msg).to.eq("Le nom d'utilisateur est requis");
            cy.contains(new RegExp(msg.replace(/[-/\\^$*+?.()|[\]{}]/g, '.'), 'i')).should('be.visible');
        });
        cy.location('pathname').should('eq', '/register');
    });

    it("affiche un message d'erreur générique en cas d'erreur serveur", () => {
        cy.intercept('POST', '**/auth/register', {
            statusCode: 500,
            headers: { 'content-type': 'application/json' },
            body: { message: "Erreur lors de la création de l'utilisateur, veuillez vérifier" },
        }).as('registerServerError');

        visitRegister();
        fillRegisterForm('Bob', 'StrongPass!');
        cy.get('button[type="submit"]').click();

        cy.wait('@registerServerError').then((interception) => {
            const msg = interception?.response?.body?.message as string;
            expect(msg).to.eq("Erreur lors de la création de l'utilisateur, veuillez vérifier");
            cy.contains(new RegExp(msg.replace(/[-/\\^$*+?.()|[\]{}]/g, '.'), 'i')).should('be.visible');
        });
        cy.location('pathname').should('eq', '/register');
    });
});

