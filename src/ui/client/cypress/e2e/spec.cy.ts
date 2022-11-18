const SCAN_TIMEOUT_MS = 30000

describe('ModronApp', () => {
  it('scans, refreshes observations', () => {
    const group = 'modron-test'
    cy.visit('/')
    cy.contains('0 observations').should('be.visible')
    cy.contains(group).parents().find('.button').first().click()
    cy.wait(2000)  // Wait for the scan to run
    cy.reload()
    cy.contains('14 observations', { timeout: SCAN_TIMEOUT_MS }).should('be.visible')
    cy.contains('Scan').should('be.visible')
    cy.contains(group).click()
  })
  it('creates exceptions', () => {
    cy.get('.notify-ctn>svg').first().should('be.visible').click()
    cy.get('textarea[formControlName="justification"]').type('trust me')
    cy.get('input[formControlName="validUntilTime"]').should(($dateTimePicker: any) => {
      let date = new Date($dateTimePicker.val())
      date.setHours(date.getHours() + 24)
      $dateTimePicker.val(date.toLocaleDateString('en-US'))
    })
    cy.get('button[type="submit"]').should('be.enabled').click()
    // Check that the exception is indeed created
    cy.get('.notify-ctn>svg').first().should('be.visible').click()
    cy.contains('trust me').should('be.visible')
  })
})
