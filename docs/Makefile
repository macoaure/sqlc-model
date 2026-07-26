IMAGE ?= python:3.12-slim
PORT ?= 8000
REPO_ROOT := $(abspath ..)

DOCKER_RUN = docker run --rm \
	-p $(PORT):8000 \
	-v "$(REPO_ROOT):/work" \
	-w /work/docs \
	$(IMAGE)

SETUP = apt-get update >/dev/null \
	&& export DEBIAN_FRONTEND=noninteractive \
	&& apt-get install -y --no-install-recommends git >/dev/null \
	&& rm -rf /var/lib/apt/lists/* \
	&& git config --global --add safe.directory /work \
	&& pip install --root-user-action=ignore -q -r requirements.txt \
	&& cp mkdocs.yml /tmp/mkdocs.yml \
	&& sed -i 's|^docs_dir: .*|docs_dir: /work/docs|; s|^site_dir: .*|site_dir: /work/site|; s|custom_dir: overrides|custom_dir: /work/docs/overrides|' /tmp/mkdocs.yml

.PHONY: serve build

serve:
	$(DOCKER_RUN) sh -c "$(SETUP) && mkdocs serve -f /tmp/mkdocs.yml -a 0.0.0.0:8000"

build:
	$(DOCKER_RUN) sh -c "$(SETUP) && mkdocs build -f /tmp/mkdocs.yml"
