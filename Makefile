test:
	@pre-commit run --all-files

install:
	@python3 -m pip install --upgrade pip setuptools
	@python3 -m pip install --upgrade -r requirements.txt

dev-install:
	@python3 -m pip install --upgrade pip setuptools
	@python3 -m pip install --upgrade -r requirements-dev.txt
	@pre-commit install
	@pre-commit autoupdate
	@pre-commit run README.md

run:
	@python3 -m alita

update:
	@git pull
	@python3 -m pip install --upgrade -r requirements.txt
