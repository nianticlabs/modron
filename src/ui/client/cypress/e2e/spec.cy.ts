const SCAN_TIMEOUT_MS = 30000

describe("ModronApp", () => {
  it("scans, refreshes observations", () => {
    const group = "modron-test"
    cy.visit("/")
    cy.contains("0 observations").should("be.visible")
    cy.contains(group).parents().find(".button").first().click()
    cy.contains("Scanning").should("be.visible")
    cy.wait(2000)  // Wait for the scan to run
    cy.contains("16 observations", { timeout: SCAN_TIMEOUT_MS }).should("be.visible")
    cy.contains("Scan").should("be.visible")
    cy.contains(group).parents().find(".resourceGroup-info").click()
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
    cy.get("div.notify-ctn").first().should("be.visible").click()
    cy.get("textarea[formControlName=\"justification\"]").type("trust me")
    cy.get("input[formControlName=\"validUntilTime\"]").should(($dateTimePicker: any) => {
      const date = new Date($dateTimePicker.val())
      date.setHours(date.getHours() + 24)
      $dateTimePicker.val(date.toLocaleDateString("en-US"))
    })
    cy.get("button[type=\"submit\"]").should("be.enabled").click()
    // Check that the exception is indeed created
    cy.get(".notify-ctn>svg").first().should("be.visible").click()
    cy.contains("trust me").should("be.visible")
  })
})
