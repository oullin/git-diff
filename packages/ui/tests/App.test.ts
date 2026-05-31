import { flushPromises, mount } from "@vue/test-utils";
import { createPinia } from "pinia";
import { describe, expect, it } from "vitest";
import App from "@entry/App.vue";
import { installBrowserFallback } from "@lib/browser-fallback";

describe("App", () => {
  it("renders local repository changes and review controls", async () => {
    installBrowserFallback();

    const wrapper = mount(App, { global: { plugins: [createPinia()] } });

    await flushPromises();

    await flushPromises();

    // Renders the local repository's changes...
    expect(wrapper.text()).toContain("src/App.vue");
    expect(wrapper.text()).toContain("changed file");
    // ...and the review controls.
    expect(wrapper.text()).toContain("Submit review");
  });
});
