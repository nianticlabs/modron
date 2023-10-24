import { defineConfig } from "cypress"

export default defineConfig({

  e2e: {
    baseUrl: "http://localhost:8080",
    supportFile: false
  },
  video: false,
  screenshotOnRunFailure: false,
  component: {
    devServer: {
      framework: "angular",
      bundler: "webpack",
    },
    specPattern: "**/*.cy.ts"
  },

  reporter: "junit",
  reporterOptions: {
    mochaFile: "/app/results/modron-e2e-ui-junit.xml",
    toConsole: false,
  },

})
