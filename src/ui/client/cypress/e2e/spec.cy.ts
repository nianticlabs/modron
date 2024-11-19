const SCAN_TIMEOUT_MS = 30000

describe("ModronApp", () => {
  it("scans, refreshes observations", () => {
    const group = "modron-test"
    cy.visit("/")
    const projectCard = cy.get(".mat-mdc-card").contains(group).parents(".mat-mdc-card")
    projectCard.contains("0 observations").should("be.visible")
    projectCard.get("mat-progress-bar").should("not.exist")

    const scanButton = cy.get(".scan-all-rgs-button");
    scanButton.should("be.visible")
    scanButton.contains("SCAN ALL")
    scanButton.click()
    projectCard.get("mat-progress-bar").should("be.visible")
    cy.get(".scan-all-rgs-button").should("be.disabled")

    cy.wait(2000)  // Wait for the scan to run

    projectCard.get(".findings-by-severity", { timeout: SCAN_TIMEOUT_MS }).should("be.visible")
    projectCard.get("mat-progress-bar").should("not.exist")
    cy.contains("SCAN").should("be.visible")

    // Iterate through the children
    projectCard.get(".findings-by-severity > div").then((elements) => {
      cy.wrap(elements.eq(0)).contains("5").should("be.visible")
      cy.wrap(elements.eq(1)).contains("14").should("be.visible")
      cy.wrap(elements.eq(2)).contains("1").should("be.visible")
    })

    projectCard.get(".mat-mdc-card").click()
    cy.contains("API_KEY_WITH_OVERBROAD_SCOPE").should("be.visible")
    cy.contains("CROSS_PROJECT_PERMISSIONS").should("be.visible")
  })
  it("visits the resource group page directly", () => {
    cy.visit("/modron/resourcegroup/projects-modron-test")
    cy.contains("API_KEY_WITH_OVERBROAD_SCOPE").should("be.visible")
    cy.contains("CROSS_PROJECT_PERMISSIONS").should("be.visible")
  })
  it("visits the stats page", () => {
    cy.visit("/modron/stats")
    cy.contains("API_KEY_WITH_OVERBROAD_SCOPE").should("be.visible")
    cy.contains("CROSS_PROJECT_PERMISSIONS").should("be.visible")
  })
  it("creates exceptions", () => {
    cy.visit("/modron/resourcegroup/projects-modron-test")
    cy.get("app-notif-bell-button").first().should("be.visible").click()
    cy.get("textarea[formControlName=\"justification\"]").type("trust me")
    cy.get("input[formControlName=\"validUntilTime\"]").should(($dateTimePicker: any) => {
      const date = new Date($dateTimePicker.val())
      date.setHours(date.getHours() + 24)
      $dateTimePicker.val(date.toLocaleDateString("en-US"))
    })
    cy.get("button[type=\"submit\"]").should("be.enabled").click()
    // Check that the exception is indeed created
    cy.get("app-notif-bell-button").first().should("be.visible").click()
    cy.contains("trust me").should("be.visible")
  })
})
