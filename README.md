# HowAreThey
A reminder system to keep in touch with your friends. Currently, this is just a backend web server with a few endpoints.
The `frontend` directory is a work in progress and has multiple bugs and not many features. The `backend` web server works like a charm but use the `frontend` at your own peril or submit some Pull Requests to tidy it up.

### Endpoints available
| Endpoint | Description |
|---|---|
| `GET /friends` | Returns a list of all the friends in the database. |
| `GET /friends/count` | Returns the number of friends in the list |
| `GET /friends/id/:id` | Returns the object with the ID specified |
| `GET /friends/name/:name` | Returns the object with the name specified |
| `GET /friends/random` | Picks a random friend from the database and returns their details |
| `DELETE /friends/:id` | Deletes the friend that matches the ID specified from the database. |
| `POST /friends` | Adds the friend using the Name, LastContacted and Notes data specified in the request. |
| `PUT /friends/:id` | Updates the friend that relates to :id specified with the new data specified in the request. |


### Docker Config
| Environment Variable | Details | Example | Default |
|---|---|---|---|
| NOTIFICATION_SERVICE | Used to define which service to use for notifications. Can be one of DISCORD, NTFY | DISCORD | N/A |
| WEBHOOK_URL | Provide a Discord webhook to send notifications to Discord. Not providing a webhook will only log the events, it won't send the notification anywhere | N/A | N/A |
| CRON | [Cron expression](https://crontab.guru/) to define how often a friend will get picked. By default, runs weekly. Use integer format for each field. | `0 0 7 * * 1` | `@weekly` |

### To run the container
Replace the tag with the tag you want to use
```
docker run -p 8080:8080 -v $PWD/sql/:/home/hat/sql/ kalmonipa/howarethey:v0.16
```

See `docker-compose.yaml` for an example Docker Compose file.
