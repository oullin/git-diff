import { createPinia, type Pinia } from "pinia";

// pinia is exported as a singleton so main.ts can install it on the root app
// before the first store call lands. Individual stores live alongside this
// file under stores/<domain>.store.ts.
export const pinia: Pinia = createPinia();
