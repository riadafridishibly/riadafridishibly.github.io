local:
	bundle exec jekyll serve --incremental

docker:
	docker compose down
	docker compose up

