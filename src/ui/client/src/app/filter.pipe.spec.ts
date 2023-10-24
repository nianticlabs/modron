import { FilterKeyValuePipe, FilterNoObservationsPipe } from "./filter.pipe";

describe("FilterKeyValuePipe", () => {
  it("create an instance", () => {
    const pipe = new FilterKeyValuePipe();
    expect(pipe).toBeTruthy();
  });
});

describe("filterNoObservations", () => {
  it("create an instance", () => {
    const pipe = new FilterNoObservationsPipe();
    expect(pipe).toBeTruthy();
  });
});
