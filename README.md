
Readme
------

This is the minimum code that seems to be necessary to connect to a
[Minecraft] server version 1.7.2 from [Mojang], but does nothing else
interesting at the moment.

It does everything that is necessary for a successful login:

1. Yggdrasil authentication and token refresh
2. Mojang session requests
3. Session key exchange via RSA
4. AES/CFB8 encryption

Note
====

This is a now just a big pile of mess and meant for personal backup mostly.

See also
========

Minecraft [Protocol]

Thanks
======

Special thanks to:

Nick Gamberini [Spock]
Carlos Cobo for [Minero]
Matthew Collins for [Netherrack]

[Protocol]: http://wiki.vg/Protocol
[Minecraft]: http://minecraft.net
[Mojang]: http://mojang.com
[Spock]: https://github.com/nickelpro/spock
[Minero]: https://github.com/toqueteos/minero
[Netherrack]: https://github.com/NetherrackDev/netherrack
