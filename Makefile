test:
	@pre-commit run --all-files

install:
	@pip3 install --upgrade pip setuptools
	@pip3 install --upgrade -r requirements.txt

dev-install:
	@pip3 install --upgrade pip setuptools
	@pip3 install --upgrade -r requirements.txt
	@pip3 install --upgrade -r requirements-dev.txt
	@pre-commit
	@pre-commit install

run:
	@python3 -m alita

update:
	@git pull
	@python3 -m pip install --upgrade -r requirements.txt
