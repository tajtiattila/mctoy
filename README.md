
Readme
------

This is the minimum code that seems to be necessary to connect to a
[Minecraft] server version 1.7.2 from [Mojang], but does not much
interesting at the moment.

Features:

* connecting servers both in online and offline mode
* respond KeepAlive messages and send player position updates to remain connected

It does everything that is necessary for a successful online login:

1. Yggdrasil authentication and token refresh
2. Mojang session requests
3. Session key exchange via RSA
4. AES/CFB8 encryption

Subpackages
===========

protocol
--------

Package protocol provides types for all packets as of 1.7.2, as well as functions to
encode and decode primitive types, types used in packets and the packets themselves.
All packets as of 1.7.2 are implemented, but some of them, especially serverbound
packets are untested.

net
---

Package net provides functions and classes for connecting and authentication.
The type ClientConn provides functions to log in and makes it easy to send and
receive protocol messages.

nbt
---

Package NBT is preliminary work on an NBT parser necessary to decode world chunk
information sent by the server.

Note
====

This is a now just a big pile of mess and meant for personal backup mostly.

See also
========

Minecraft [Protocol]

Thanks
======

Special thanks to:

* Nick Gamberini for [Spock]
* Carlos Cobo for [Minero]
* Matthew Collins for [Netherrack]

[Protocol]: http://wiki.vg/Protocol
[Minecraft]: http://minecraft.net
[Mojang]: http://mojang.com
[Spock]: https://github.com/nickelpro/spock
[Minero]: https://github.com/toqueteos/minero
[Netherrack]: https://github.com/NetherrackDev/netherrack
