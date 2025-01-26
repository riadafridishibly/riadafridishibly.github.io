.PHONY: local
local:
	bundle exec jekyll serve --incremental

.PHONY: docker
docker:
	docker compose down
	docker compose up

.PHONY: clean
clean:
	rm -rf _site/
