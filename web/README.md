## Decap Web Service

This is the core web service for Decap.  It exposes REST API endoints
for a web UI frontend, and receives post commit hooks from Github
and Atlassian Stash, and parses those events and launches a new
build in the build container accordingly.

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
        <td>Stash</td>
        <td>/hooks/stash</td>
    </tr>
</table>

### Manually Launching Builds

TBD
