import { flushPromises, mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import App from "@entry/App.vue";
import { installBrowserFallback } from "@lib/browser-fallback";

describe("App", () => {
    it("renders local repository changes and review controls", async () => {
        installBrowserFallback();
        const wrapper = mount(App);
        await flushPromises();
        await flushPromises();

        expect(wrapper.text()).toContain("/Users/local/project");
        expect(wrapper.text()).toContain("src/App.vue");
        expect(wrapper.text()).toContain("Reviews");
    });
});
