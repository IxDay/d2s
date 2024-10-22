Day 2 Stack
===========

D2S (Day 2 Stack) is a Golang skeleton app built with production and Kubernetes readiness in mind.

Why D2S?
--------

Day-2 operations involve the ongoing maintenance, monitoring, and optimization of your
deployed applications (see [this article for details][day_2_ref]).
In today’s fast-paced development environment, frameworks often prioritize DX and rapid setup.
But when it comes to scaling for production, they may require significant adjustments.
D2S addresses this by offering a production-focused skeleton app from the start,
tailored for Kubernetes and real-world deployment scenarios.
We believe you shouldn't have to choose between a good DX and being production-ready.

Tech Stack
----------

D2S is currently built with the following technologies:

- Golang: Reliable and performant backend.
- Templ: HTML templates for fast, server-rendered pages.
- HTMX: Modern approach for front-end interactivity without heavy JavaScript.
- Tailwind: Utility-first CSS framework for rapid and customizable UI development.


Features and Roadmap
--------------------

We are actively developing D2S, and here’s a snapshot of what's included and what’s coming:

- [ ] DB connection + Routing + Rendering (the basic hello world feature)
	- [x] Application leverage HTMX to demo partial loading when navigating between pages.
	- [x] Live reloading on code changes
- [x] 12 factor app:
	- [x] Configuration is loaded as follow:
		```
		Files < Env variables < CLI args
		```
	- [x] Logs are outputed to stdout
- [ ] DB operations
	- [ ] Migrations
	- [ ] Backup + Restore + Test
- [ ] Kubernetes ready:
	- [ ] Containerized
	- [x] Liveness/Readiness probes
	- [x] Graceful shutdown
- [x] Monitoring:
	- [x] Logs
		- [x] Structured
		- [x] Stack trace
		- [x] Tracing correlation
	- [x] Metrics
	- [x] Tracing
- [x] Caching
- [ ] Security
- [ ] CI/CD
	- [ ] Image build with caching
	- [x] Additional file format checks (`editorconfig`, `shellcheck`)
	- [ ] Unit test
	- [ ] Integration test?
	- [x] Vulnerability checking

**Bonus:** Application can be run locally without going through a container.
This ease the debugging process by simplifying the build phase
(no build container + remote debug anymore)


Design Decisions
----------------

We believe that transparency in technology choices is crucial. 
Here’s why we opted for each key component:

### Base technologies

- Golang: Chosen for its simplicity, concurrency model, and strong performance in production environments.
- Templ: Provides an efficient templating solution that avoids unnecessary complexity while fitting well into the Golang ecosystem.
- HTMX: Allows for dynamic front-end behavior without a full SPA framework, offering flexibility with minimal JavaScript.
- Tailwind: Chosen for its productivity boost and design consistency, letting developers build clean UIs quickly.

### Project structure

While the structure is not final, some decisions have already been made.
The following artifacts influenced these decisions: [this documentation entry][layout_doc] and [this issue][layout_issue].

Since the main entry point is the launch of the server binary, a `main.go` file is located at the root of this repository. 
It also includes a configuration example (`d2s.example.toml`), 
which can be copied and will be loaded by default if name is `d2s.toml`.

The app directory follows a structure inspired by the Pages Router from Next.js. 
Though it lacks the hidden logic of Next.js, this layout is a pragmatic way to 
represent the website's arborescence in the file system.

The internal directory is currently too large and will need to be broken down into smaller parts.

### Libraries

#### MRuby and MRake

#### Zerolog

*What about Slog and Zap?*

#### Chi 

#### Opentelemetry

#### Viper + Cobra

#### Prometheus

#### Xerrors

#### Air


Getting started
---------------

In order to start using this skeleton you will need to install some binaries

```sh
go install github.com/a-h/templ/cmd/templ@v0.2.778
go install github.com/air-verse/air@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

Contributing
------------

Contributions are welcome! Please check out the roadmap and open issues if you want to get involved, 
and feel free to submit PRs.

License
-------

This project is licensed under the MIT License.

[day_2_ref]: https://www.qovery.com/blog/day-0-day-1-day-2-what-are-the-differences/
[layout_doc]: https://go.dev/doc/modules/layout
[layout_issue]: https://github.com/golang-standards/project-layout/issues/117
[nextjs_pages_router]: https://nextjs.org/docs/pages/building-your-application/routing/pages-and-layouts
