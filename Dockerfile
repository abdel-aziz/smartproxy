# Blueboard backoffice api

# VERSION       0.1

FROM buildpack-deps:jessie

MAINTAINER Francis Bouvier, francis@blueboard.it

COPY bb_proxy /usr/local/bin
COPY proxies.txt /

EXPOSE 9999

CMD ["bb_proxy", "-D", "--tor", "tor:9050", "--addr", "0.0.0.0:9999"]
