test:
	@pre-commit run --all-files

install:
	@pip3 install --upgrade pip setuptools wheel poetry
	@sleep 3
	@poetry config virtualenvs.create false
	@sleep 3
	@poetry install --no-dev --no-interaction

run:
	@python3 -m alita

clean:
	@rm -rf alita/logs
	@pyclean .
