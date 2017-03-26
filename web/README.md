## Decap Web Service

This project's dependencies are managed by the
[dep](https://github.com/golang/dep) tool, which is emerging as the
official Go dependency management tool.

This is the core web service for Decap.  It exposes REST API endoints
for a web UI frontend, and receives post commit hooks from Github
(and eventually other repository managers, including Stash and
Bitbucket), and parses those events and launches a new build in the
build container accordingly.

A given branch on a project is locked in etcd to ensure that only
one container is building this project branch at any one time.  This
may be more important for some projects types than others.

### Post commit hook endpoints

HTTP endpoints for various source code management systems

<table>
    <tr>
        <th>Repository Manager Name</th>
        <th>URL</th>
    </tr>
    <tr>
        <td>Github</td>
        <td>/hooks/github</td>
    </tr>
    <tr>
        <td>Build scripts repository reload</td>
        <td>/hooks/buildscripts</td>
    </tr>
</table>

Github post-commit hooks should be pointed at _/hooks/github_.
Decap will parse the payload and launch a build accordingly.

The _/hooks/buildscripts_ endpoint is a special endpoint that
post-commit hooks on the build-scripts repository should hit.  It
forces a reload of the managed Projects.

### REST API

Read more about the Decap [REST API](https://github.com/ae6rt/decap/wiki/API).

### More information

See the [Developer Getting Started Guide for more information](https://github.com/ae6rt/decap/wiki/Developing)
