FROM angular_io_base

ADD repo /opt/angular
ADD repo /opt/repo
WORKDIR /opt/repo
RUN npm install
RUN gulp build-compile

RUN rm -rf /usr/local/apache2/htdocs && \
    ln -s /opt/repo/www /usr/local/apache2/htdocs
