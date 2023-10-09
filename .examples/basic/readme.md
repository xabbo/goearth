# Basic

This is an example extension that demonstrates basic event handling, packet manipulation, sending and receiving.\
It handles and logs the following events:

* Extension initialization
  * Whether a game connection has already been established
* Extension activation
  * This is when the user clicks the green "play" button in G-Earth.
  * An outgoing `RetrieveInfo` message is sent to get the current user's data.
  * A `UserObject` packet is received and the user's ID and name are output.
* Game connection
  * The remote host and port
  * The client identifier and version
  * The number of message infos received
* Intercept all messages
  * Prints if the packet is a `Ping` message
* Intercept outgoing `Chat`, `Shout` and `Whisper` messages
  * Reads the content of the chat message
  * Replaces the word `"apple"` with `"orange"`
  * Blocks the message if it contains the word `"block"`
* Game disconnection