# HowAreThey
A reminder system to keep in touch with your friends

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

### Example docker-compose.yaml
```
version: '3'

services:
  howarethey:
    container_name: howarethey
    image: kalmonipa/howarethey:v0.16
    restart: always
    environment:
        # Defines which service to send notifications too
      - NOTIFICATION_SERVICE=DISCORD
        # Discord Webhook to send notifications too. Read the discord webhook docs to get one
      - WEBHOOK_URL=https://discord.com/api/webhooks/myexamplewebhook
        # Picks a friend at 7am every Monday UTC
      - CRON=0 0 7 * * 1
    ports:
      # The Web UI (enabled by --api.insecure=true)
      - "8022:8080"
    volumes:
        # Define where your persistent storage goes too
      - path/to/sql/dir/:/home/hat/sql
```
