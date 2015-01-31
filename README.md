# Alexandria CMDB Dashboard

*A CMDB from the future!!!*

Alexandria CMDB is an open source configuration management database written in [Google Go](https://golang.org/) with a [MongoDB](http://www.mongodb.org/) backend.

This project is in infancy and not ready for deployment. It aims to achieve the following:

* Fast, lightweight and low configuration overhead
* Intuitive and responsive frontend
* Automated data sourcing, transformation and validation
* Vertical and horizontal scalability
* High availability
* Message queue integration
* Comprehensive RESTful API
* Multitenanted, cloud or on-premise
* ITIL friendly

For more information, see the [Alexandria CMDB docs](http://cavaliercoder.github.io/alexandria-docs/).

## License

Alexandria CMDB - Open source configuration management database
Copyright (C) 2014  Ryan Armstrong (ryan@cavaliercoder.com)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
    
## Libraries

* [Revel](http://revel.github.io/index.html)

* [SB Admin](http://startbootstrap.com/template-overviews/sb-admin/)

### Start the web server:

  `revel run github.com/cavaliercoder/alexandria-dashboard`

   Run with <tt>--help</tt> for options.

### Go to http://localhost:9000/

### Description of Contents

The default directory structure of a generated Revel application:

    myapp               App root
      app               App sources
        controllers     App controllers
          init.go       Interceptor registration
        models          App domain models
        routes          Reverse routes (generated code)
        views           Templates
      tests             Test suites
      conf              Configuration files
        app.conf        Main configuration file
        routes          Routes definition
      messages          Message files
      public            Public assets
        css             CSS files
        js              Javascript files
        images          Image files

app

    The app directory contains the source code and templates for your application.

conf

    The conf directory contains the application’s configuration files. There are two main configuration files:

    * app.conf, the main configuration file for the application, which contains standard configuration parameters
    * routes, the routes definition file.


messages

    The messages directory contains all localized message files.

public

    Resources stored in the public directory are static assets that are served directly by the Web server. Typically it is split into three standard sub-directories for images, CSS stylesheets and JavaScript files.

    The names of these directories may be anything; the developer need only update the routes.

test

    Tests are kept in the tests directory. Revel provides a testing framework that makes it easy to write and run functional tests against your application.

### Follow the guidelines to start developing Alexandria CMDB Dashboard:

* The [Getting Started with Revel](http://revel.github.io/tutorial/index.html).
* The [Revel guides](http://revel.github.io/manual/index.html).
* The [Revel sample apps](http://revel.github.io/samples/index.html).
* The [API documentation](http://revel.github.io/docs/godoc/index.html).
