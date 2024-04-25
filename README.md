# HowAreThey
A reminder system to keep in touch with your friends. It's a simple tool to prompt you to get in touch with people that you might lose touch with over time.

If you define a notification service and pass in a webhook URL, it will send a message to your channel/device. Currently it only supports Discord and Ntfy.
The message will look something like
```
HowAreThey
You should get in touch with Jack Reacher. You haven't spoken to them since 06/06/2023.
```

Currently, this is just intended as a backend web server with a few endpoints, storing the data in a SQLite database. Updating the info is a bit finnicky if you're not used to using CLI/Curl/etc.

I'm also working on a frontend web page for this but Javascript is hard and don't currently have the time to learn it at the moment.

### Usage
#### To run the container
Replace the tag with the tag you want to use
```
docker run -p 8080:8080 -v $PWD/sql/:/home/hat/sql/ kalmonipa/howarethey:v0.16
```

See `docker-compose.yaml` for an example Docker Compose file.

#### Examples

Here is a JSON object of an example person:
```
{
  "ID": "1",
  "Name": "Steve Carell",
  "LastContacted": "06/06/2023",
  "Birthday": "16/08/1962",
  "Notes": "Ask him how his store is going in Marshfield"
}
```

To update the notes for example, you can send a request using the `/friends/:id` endpoint. Only the keys that are provided will get updated.
```
curl "http://localhost:8080/friends/1" \
    --request PUT \
    --header "Content-Type: application/json" \
    --data "{\"Notes\":\"His store is going great\"}"
```
To update both the Notes and the LastContacted field, you would send something like this
```
curl "http://localhost:8080/friends/1" \
        --request PUT \
        --header "Content-Type: application/json" \
        --data "{\"LastContacted\":\"17/04/2024\",\"Notes\":\"His store is going great\"}"
```

Calling `GET /friends/random` will trigger a random friend to get chosen, their `LastContacted` field to get updated to today and a notification will get sent to your notification service specified in the env var (if any is set)


### Endpoints available
| Endpoint | Description |
|---|---|
| `GET /friends` | Returns a list of all the friends in the database. |
| `GET /birthdays` | Returns a list of all the friends that have birthdays today |
| `GET /friends/count` | Returns the number of friends in the list |
| `GET /friends/id/:id` | Returns the object with the ID specified |
| `GET /friends/name/:name` | Returns the object with the name specified |
| `GET /friends/random` | Picks a random friend from the database and returns their details |
| `DELETE /friends/:id` | Deletes the friend that matches the ID specified from the database. |
| `POST /friends` | Adds the friend using the Name, LastContacted, Birthday and Notes data specified in the request. |
| `PUT /friends/:id` | Updates the friend that relates to :id specified with the new data specified in the request. |


### Docker Config
| Environment Variable | Details | Example | Default |
|---|---|---|---|
| NOTIFICATION_SERVICE | Used to define which service to use for notifications. Can be one of DISCORD, NTFY | DISCORD | N/A |
| WEBHOOK_URL | Provide a Discord webhook to send notifications to Discord. Not providing a webhook will only log the events, it won't send the notification anywhere | N/A | N/A |
| FRIEND_SELECTOR_CRON_SCHEDULE | [Cron expression](https://crontab.guru/) to define how often a friend will get picked. By default, runs weekly. Use integer format for each field. | `0 7 * * 1` | `@weekly` |
| BIRTHDAY_CHECK_TIME | What time of day the app should check for birthdays. Must be within 0-23; 0 being midnight-1am, 23 being 11pm-midnight | `"8"` | `8` |
| IGNORE_BIRTHDAYS | Set to `true` if you don't want the app to check for birthdays | `true` | `false` |

### Development
Write any new tests and run the following commands from the root directory
#### Unit tests
`go test -v ./pkg/test/unit_test`

#### Integration tests
`./run-integration-tests.sh`
To skip the image build step (if you already have an image for your feature branch on your local machine), set the `BUILD_IMAGE` env var to `false` and run the int test script. i.e:
`BUILD_IMAGE="false" ./run-integration-tests.sh`

If the tests pass submit a Pull Request with a detailed description.
