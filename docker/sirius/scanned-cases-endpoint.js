const { respond, stores } = require("@imposter-js/types");

const lpaIdStore = stores.open("lpaIdStore");

const lpaIdCounter = lpaIdStore.load("counter") ?? 100;

respond()
  .withStatusCode(201)
  .withContent(`{"uid":"${700000000000 + lpaIdCounter}"}`);

lpaIdStore.save("counter", lpaIdCounter + 1);
