import { expect, test } from "@oclif/test";
import * as nock from "nock";
import * as sinon from "sinon";

const PAGE_ID = "12345678-1234-1234-1234-123456789012";
const PAGE_ID_NO_DASHES = PAGE_ID.replace(/-/g, "");

describe("page:retrieve:property_item", () => {
  let processExitStub: sinon.SinonStub;

  beforeEach(() => {
    nock.cleanAll();
    processExitStub = sinon.stub(process, "exit" as any);
  });

  afterEach(() => {
    nock.cleanAll();
    processExitStub.restore();
  });

  const response = {
    object: "list",
    results: [
      {
        type: "title",
        id: "title",
        title: {
          type: "title",
          text: {
            content: "dummy title",
            link: null,
          },
        },
      },
    ],
    next_cursor: null,
    has_more: false,
    type: "property_item",
  };

  test
    .do(() => {
      nock("https://api.notion.com")
        .get(`/v1/pages/${PAGE_ID_NO_DASHES}/properties/dummy-property-id`)
        .reply(200, response);
    })
    .stdout()
    .command(["page:retrieve:property_item", PAGE_ID, "dummy-property-id"])
    .it("shows retrieved page object when success", (ctx) => {
      expect(ctx.stdout).to.contain('object": "list');
      expect(ctx.stdout).to.contain('type": "property_item');
    });
});
