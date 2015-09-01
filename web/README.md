## Aftomato Web Service

This is the core web service for Aftomato.  It receives post commit
hooks from Stash, Github, and Bitbucket, parses those events and
launches a new build in the build container accordingly.

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
    <tr>
        <td>Bitbucket</td>
        <td>/hooks/bitbucket</td>
</table>

### Manually Launching Builds

TBD
