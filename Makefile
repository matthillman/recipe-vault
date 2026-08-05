SHELL := /bin/sh

GO ?= go
SITE_DIR := site
CARDS_OUT ?= out/cards.pdf
RECIPES ?=

.PHONY: check validate-recipes go-test scripts-test recipe-import cards-list cards-build site-sync site-dev site-build site-test

check: validate-recipes go-test scripts-test site-test

validate-recipes:
	sh ./scripts/validate-recipes.sh

go-test:
	@if command -v $(GO) >/dev/null 2>&1; then \
		GOCACHE=/tmp/gocache $(GO) test ./...; \
	else \
		echo "go not installed; skipping go tests"; \
	fi

scripts-test:
	@if command -v node >/dev/null 2>&1; then \
		node --test scripts/decode-import-issue.test.mjs; \
	else \
		echo "node not installed; skipping script tests"; \
	fi

recipe-import:
	@if [ -z "$(URL)" ] && [ -z "$(SOURCE_TEXT_FILE)" ]; then \
		echo "set URL=https://... and/or SOURCE_TEXT_FILE=/path/to/recipe.txt"; \
		exit 2; \
	fi
	@if command -v $(GO) >/dev/null 2>&1; then \
		GOCACHE=/tmp/gocache $(GO) run ./cmd/recipeimport import \
			$(if $(URL),--url "$(URL)",) \
			$(if $(SOURCE_TEXT_FILE),--source-text-file "$(SOURCE_TEXT_FILE)",) \
			--out recipes; \
	else \
		echo "go not installed; cannot import recipes"; \
		exit 1; \
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

site-test:
	@if command -v npm >/dev/null 2>&1; then \
		cd $(SITE_DIR) && npm test; \
	else \
		echo "npm not installed; skipping site tests"; \
	fi
