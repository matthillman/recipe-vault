SHELL := /bin/sh

GO ?= go
SITE_DIR := site
CARDS_OUT ?= out/cards.pdf
RECIPES ?=

.PHONY: check validate-recipes go-test cards-list cards-build site-sync site-dev site-build

check: validate-recipes go-test

validate-recipes:
	sh ./scripts/validate-recipes.sh

go-test:
	@if command -v $(GO) >/dev/null 2>&1; then \
		GOCACHE=/tmp/gocache $(GO) test ./...; \
	else \
		echo "go not installed; skipping go tests"; \
	fi

cards-list:
	@if command -v $(GO) >/dev/null 2>&1; then \
		GOCACHE=/tmp/gocache $(GO) run ./cmd/recipecards list; \
	else \
		echo "go not installed; cannot list recipe cards"; \
		exit 1; \
	fi

cards-build:
	@if [ -z "$(RECIPES)" ]; then \
		echo "set RECIPES=slug1,slug2"; \
		exit 2; \
	fi
	@if command -v $(GO) >/dev/null 2>&1; then \
		GOCACHE=/tmp/gocache $(GO) run ./cmd/recipecards build --out "$(CARDS_OUT)" --recipes "$(RECIPES)"; \
	else \
		echo "go not installed; cannot build recipe cards"; \
		exit 1; \
	fi

site-sync:
	cd $(SITE_DIR) && npm run sync

site-dev:
	cd $(SITE_DIR) && npm run dev

site-build:
	cd $(SITE_DIR) && npm run build
