(() => {
	"use strict";

	const themeStorageKey = "shiftbell-theme";
	const themePreference = window.matchMedia("(prefers-color-scheme: dark)");
	let dialogTrigger = null;
	let completionPendingTrigger = null;

	const isTheme = (value) => value === "light" || value === "dark";

	const readStoredTheme = () => {
		try {
			const theme = localStorage.getItem(themeStorageKey);
			return isTheme(theme) ? theme : null;
		} catch {
			return null;
		}
	};

	const storeTheme = (theme) => {
		try {
			localStorage.setItem(themeStorageKey, theme);
		} catch {
			return;
		}
	};

	const preferredTheme = () => themePreference.matches ? "dark" : "light";

	const syncThemeControls = (theme) => {
		const nextTheme = theme === "dark" ? "light" : "dark";
		const label = `Switch to ${nextTheme} theme`;

		document.querySelectorAll("[data-theme-toggle]").forEach((button) => {
			button.setAttribute("aria-label", label);
			button.setAttribute("title", label);
			const text = button.querySelector("[data-theme-label]");
			if (text) {
				text.textContent = label;
			}
		});
	};

	const applyTheme = (theme) => {
		document.documentElement.setAttribute("data-bs-theme", theme);
		syncThemeControls(theme);
	};

	applyTheme(readStoredTheme() ?? preferredTheme());

	const restoreDialogFocus = () => {
		const trigger = dialogTrigger;
		dialogTrigger = null;
		if (trigger?.isConnected) {
			trigger.focus({ preventScroll: true });
		}
	};

	const removeDialog = () => {
		document.getElementById("dialog-root")?.replaceChildren();
		restoreDialogFocus();
	};

	const closeDialog = () => {
		const dialog = document.querySelector("#dialog-root dialog");
		if (dialog?.open) {
			dialog.close();
			return;
		}
		removeDialog();
	};

	const showDialog = (root) => {
		const dialog = root.querySelector("dialog");
		if (!dialog || typeof dialog.showModal !== "function") {
			restoreDialogFocus();
			return;
		}
		dialog.addEventListener("close", removeDialog, { once: true });
		dialog.addEventListener("cancel", (event) => {
			if (dialog.querySelector('[data-completion-form][aria-busy="true"]')) {
				event.preventDefault();
			}
		});
		dialog.showModal();
	};

	const clearCompletionPendingTrigger = () => {
		completionPendingTrigger?.removeAttribute("data-completion-pending");
		completionPendingTrigger = null;
	};

	const htmxRequestElement = (event) => event.detail.requestConfig?.elt ?? event.detail.elt;

	const setCompletionPending = (form, pending) => {
		const button = form.querySelector("[data-complete-action]");
		const cancelButton = form.querySelector("[data-dialog-cancel]");
		const label = button?.querySelector("[data-complete-label]");
		const dialog = form.closest("dialog");
		if (!button || !cancelButton || !label) {
			return;
		}

		if (pending && document.activeElement === button) {
			dialog?.focus({ preventScroll: true });
		}
		button.disabled = pending;
		cancelButton.disabled = pending;
		label.textContent = pending ? "Completing\u2026" : "Complete chore";
		if (pending) {
			button.setAttribute("aria-busy", "true");
			form.setAttribute("aria-busy", "true");
			clearCompletionPendingTrigger();
			completionPendingTrigger = dialogTrigger;
			completionPendingTrigger?.setAttribute("data-completion-pending", "");
			return;
		}

		button.removeAttribute("aria-busy");
		form.removeAttribute("aria-busy");
		clearCompletionPendingTrigger();
		if (
			dialog?.isConnected &&
			(document.activeElement === dialog || !dialog.contains(document.activeElement))
		) {
			button.focus({ preventScroll: true });
		}
	};

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

	document.addEventListener("click", (event) => {
		const completeButton = event.target.closest?.("[data-complete-chore]");
		if (completeButton) {
			dialogTrigger = completeButton;
		}

		if (event.target.closest?.("[data-dialog-cancel]")) {
			event.preventDefault();
			closeDialog();
			return;
		}

		const button = event.target.closest?.("[data-theme-toggle]");
		if (!button) {
			return;
		}

		const theme = document.documentElement.getAttribute("data-bs-theme");
		const nextTheme = theme === "dark" ? "light" : "dark";
		applyTheme(nextTheme);
		storeTheme(nextTheme);
	});

	themePreference.addEventListener("change", (event) => {
		if (readStoredTheme() === null) {
			applyTheme(event.matches ? "dark" : "light");
		}
	});

	window.addEventListener("storage", (event) => {
		if (event.key === themeStorageKey) {
			applyTheme(isTheme(event.newValue) ? event.newValue : preferredTheme());
		}
	});

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

	document.addEventListener("htmx:beforeRequest", (event) => {
		const form = htmxRequestElement(event)?.closest?.("[data-completion-form]");
		if (form) {
			setCompletionPending(form, true);
		}
	});

	document.addEventListener("htmx:afterRequest", (event) => {
		const requestElement = htmxRequestElement(event);
		const form = requestElement?.closest?.("[data-completion-form]");
		if (form && !event.detail.successful) {
			setCompletionPending(form, false);
		}
		if (requestElement?.matches?.("[data-chore-collection]")) {
			clearCompletionPendingTrigger();
		}
	});

	document.addEventListener("htmx:afterSwap", (event) => {
		syncNavigation();
		syncThemeControls(document.documentElement.getAttribute("data-bs-theme"));

		if (event.detail.target?.id === "dialog-root") {
			showDialog(event.detail.target);
		}

		if (event.detail.target?.id === "main") {
			closeNavigation();
			document.getElementById("main")?.focus({ preventScroll: true });
		}
	});

	document.addEventListener("choreCompleted", closeDialog);

	window.addEventListener("popstate", syncNavigation);
	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", () => {
			syncNavigation();
			syncThemeControls(document.documentElement.getAttribute("data-bs-theme"));
		}, { once: true });
	} else {
		syncNavigation();
		syncThemeControls(document.documentElement.getAttribute("data-bs-theme"));
	}
})();
