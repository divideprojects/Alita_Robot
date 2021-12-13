FROM ghcr.io/divideprojects/docker-python-base:latest AS build

# Build virtualenv as separate step: Only re-execute this step when pyproject.toml or poetry.lock changes
FROM build AS build-venv
COPY pyproject.toml poetry.lock /
RUN /venv/bin/poetry export -f requirements.txt --without-hashes --output requirements.txt
RUN /venv/bin/pip install --disable-pip-version-check -r /requirements.txt

# Copy the virtualenv into a distroless image
FROM gcr.io/distroless/python3-debian11
WORKDIR /app
COPY --from=build-venv /venv /venv
COPY . .
ENTRYPOINT ["/venv/bin/python3"]
CMD ["-m", "alita"]
