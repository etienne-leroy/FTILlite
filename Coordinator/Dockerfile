# syntax=docker/dockerfile:1
FROM python:3.8.13-bullseye as build

COPY requirements.txt /tmp/requirements.txt
RUN pip install -r /tmp/requirements.txt
COPY ftillitelibrary /tmp/ftillitelibrary
RUN pip install /tmp/ftillitelibrary
WORKDIR /app
# Establish the runtime user (with no password and no sudo)
# RUN useradd -m myapp
# USER myapp
# Run server

FROM build as publish

VOLUME /app/notebooks

CMD ["jupyter", "lab","--ip=0.0.0.0", "--port=3000","--allow-root","--no-browser", "--NotebookApp.token='ftillite'"]