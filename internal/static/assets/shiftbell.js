(() => {
	"use strict";

	const syncNavigation = () => {
		const path = window.location.pathname;
		const picker = path === "/chore-templates" && new URLSearchParams(window.location.search).get("picker") === "1";

		document.querySelectorAll("[data-shiftbell-nav]").forEach((link) => {
			const href = new URL(link.href, window.location.href).pathname;
			const activePath = picker ? "/chores" : path;
			const active = activePath === href || activePath.startsWith(`${href}/`);

			link.classList.toggle("active", active);
			if (active) {
				link.setAttribute("aria-current", "page");
			} else {
				link.removeAttribute("aria-current");
			}
		});
	};

	const closeNavigation = () => {
		const navigation = document.getElementById("primary-navigation");
		const Collapse = globalThis.bootstrap?.Collapse;
		if (!navigation?.classList.contains("show") || !Collapse) {
			return;
		}

		Collapse.getOrCreateInstance(navigation).hide();
	};

	document.addEventListener("htmx:beforeSwap", (event) => {
		const status = event.detail.xhr.status;

		if (status >= 400 && status < 600) {
			event.detail.shouldSwap = true;
		}

	});

	document.addEventListener("htmx:afterSwap", (event) => {
		syncNavigation();

		if (event.detail.target?.id === "main") {
			closeNavigation();
			document.getElementById("main")?.focus({ preventScroll: true });
		}
	});

	window.addEventListener("popstate", syncNavigation);
	syncNavigation();
})();
