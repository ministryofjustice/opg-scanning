const { respond, stores, context } = require("@imposter-js/types");

const lpaIdStore = stores.open("lpaIdStore");

const body = JSON.parse(context.request.body);

if (body.batchId == 'bad-batch') {
  respond()
    .withStatusCode(400)
    .withContent('{}');
} else {
  const lpaIdCounter = lpaIdStore.load("counter") ?? 100;

  respond()
    .withStatusCode(201)
    .withContent(`{"uid":"${700000000000 + lpaIdCounter}"}`);

  lpaIdStore.save("counter", lpaIdCounter + 1);
}
