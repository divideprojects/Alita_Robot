FROM python:3.8.6-slim-buster

COPY . /root

WORKDIR /root

RUN python3 -m pip install --upgrade pip

RUN python3 -m pip install --no-cache-dir -r ./requirements.txt

CMD ["python3" ,"-m", "alita"]