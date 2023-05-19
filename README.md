# This project is now obsoleted in favour of [serverless-scheduler](https://github.com/dezeroku/serverless-scheduler)

# Page Monitor (monitor-page)
Self hosted k8s solution to monitor websites for changes in either HTML or screenshots of them.
When such a change is detected user is notified via mail.

## Why?
I've written it for myself and use it for monitoring couple pages, but the main idea behind it was to get a practical grasp on the microservices.
That's why the components are split as much as possible.
The better approach would be probably to just use a message-broker such as `RabbitMQ` instead of running a new checker instance for every URL.

## How does it look?
Couple screenshots from the `front` provided UI.

![Login](docs/pictures/login.png?raw=true "Login")
![Home Page (empty)](docs/pictures/home.png?raw=true "Home Page (empty)")
![Task Creation](docs/pictures/create.png?raw=true "Task Creation")
![Home Page](docs/pictures/added.png?raw=true "Home Page")
![Modify Modal](docs/pictures/modify.png?raw=true "Modify Modal")

## Structure
Let me shed some light on the general design and describe the more important ones.
Each of the modules listed below is stored in a separate directory with its own README listing possible configuration options.
If the component exposes an API, there is a swagger file describing it in the appropriate directory.
The only components that are meant to be exposed to the outside world at any point are `manager` and `front`.

At the moment, the whole deployment consists of few modules that are required, such as:
* manager
* sender

and few optional ones:
* front
* db (this is a rather special case in terms of "being optional", you are still going to have the DB if you reject this, but it's going to be sqlite3 local one)
* screenshoter
* comparator

Most of the modules are accessible via the REST API (except for front and db, these are special).

### Manager
That's the core part of the system.
It's meant to keep the user list and control (spin up and terminate) the `checker` deployments.
It's possible to use it with the local `sqlite3` db or connect to `postgres` (the so-called `db` module).

### Sender
Responsible for contacting with the outside world.
For now only one API call that matters (`mail`), that uses the `sendgrid` API under the hood.

### Front
UI that contacts with the `manager` over the API.
That's the entry point, if you don't want to use the API directly.

### Screenshoter
Takes a screenshot of the provided URL.
Under the hood it's using `puppeteer`.

### Comparator
Provided with two images it calculates the structural similarity (basically a diff value) between these two.
It's used to mark the borders of what changed between the checks.

### Checker
Shouldn't be used directly at any point.
It's the smallest component actually responsible for calling the other services.
It's gives a single URL to check and periodically compares an HTML diff between the previous and current content.
In case of any difference, it sends out a mail informing about the change.


## Deployment
The easiest way to deploy is to use the generated helm package (`k8s/monitor-page`, it's also uploaded as the github package).
The minimum set of values to be overriden is provided in `k8s/monitor-page/override-example.yaml`.
Values in that file have are specific for basically every other deployment.
Requires the [`local-path-provisioner`](https://github.com/rancher/local-path-provisioner) operator to be installed in cluster.

You can run `helm package` in the `k8s/monitor-page` directory to create the package or get it from the "Actions" page from the specific commit.


## Building
All of the images are provided under the `ghcr.io/dezeroku` namespace and are automatically rebuilt on every change in the `master` branch of the repo.

Each of the components can be built via the provided Dockerfile.
To get the production build:
```
# Skip the test stages with proper runner
export DOCKER_BUILDKIT=1
docker build -t production .
```

To run tests:
```
docker build -t test . --target test
docker run -it test
```
