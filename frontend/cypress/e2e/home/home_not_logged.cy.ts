/// <reference types="cypress" />

describe("Accueil — utilisateur non connecté", () => {
  const baseUrl = Cypress.config('baseUrl') || 'http://localhost:3000';

  const closeNewsBannerIfPresent = () => {
    cy.get('body').then(($body) => {
      const closeBtn = $body.find('button[aria-label="Fermer"]');
      if (closeBtn.length) {
        cy.wrap(closeBtn.first()).click({ force: true });
      }
    });
  };

  beforeEach(() => {
    cy.clearLocalStorage();
    // Par sécurité, stub GET /game pour éviter effets secondaires si appelé
    cy.intercept('GET', '**/game', {
      statusCode: 200,
      headers: { 'content-type': 'application/json' },
      body: { games: [] },
    }).as('getGames');
  });

  it("affiche le hero et les CTA Connexion/Inscription", () => {
    cy.visit(`${baseUrl}/`);
    // si un call part malgré l'absence de token
    cy.wait(100);
    closeNewsBannerIfPresent();

    cy.contains('h1', 'Bienvenue sur Scrabble en ligne').should('be.visible');
    cy.contains('p', 'Jouez au Scrabble en ligne').should('be.visible');
    cy.contains('a', 'Connexion').should('have.attr', 'href', '/login');
    cy.contains('a', 'Inscription').should('have.attr', 'href', '/register');
  });

  it("navigue vers /login puis /register via les boutons", () => {
    cy.visit(`${baseUrl}/`);
    closeNewsBannerIfPresent();

    cy.contains('a', 'Connexion').click();
    cy.location('pathname').should('eq', '/login');

    cy.go('back');
    closeNewsBannerIfPresent();
    cy.contains('a', 'Inscription').click();
    cy.location('pathname').should('eq', '/register');
  });
});
