# God of War game browser
Tool helps to investigate files formats of the game.
Based on version: *God of War (2005)  NTSC(USA)  PS2DVD-9*

## Instruction
- Install golang compiler ([official instruction](https://golang.org/doc/install))
- `go get -u github.com/mogaika/god_of_war_browser` (this command download code and dependencies)
- Go to directory with required [tool](https://github.com/mogaika/god_of_war_browser/tree/dev/tools) and run `go build` to compile
- Do not forget read README
- If you encounter any problem or have a question - welcome to our friendly [gitter chat](https://gitter.im/god_of_war/)

## What if I want to mod game?
You can! But it is hard at this point :(
- Download required wads using god_of_war_browser web interface
- Use [wadunpack](https://github.com/mogaika/god_of_war_browser/tree/dev/tools/wadunpack) to unpack wad where you want to make change
- Add/remove/modify game files, do not forget to modify *wad_meta.txt*. Some info about file formats you can find in [sources](https://github.com/mogaika/god_of_war_browser/tree/dev/pack/wad) of god_of_war_browser.
- Archive files to wad using [wadpack](https://github.com/mogaika/god_of_war_browser/tree/dev/tools/wadpack)
- Pack wad files in part1.pak using [packer](https://github.com/mogaika/god_of_war_browser/tree/dev/tools/packer)
- For test replace original part1.pak and godofwar.toc in your iso using [isoreplacer](https://github.com/mogaika/god_of_war_browser/tree/dev/tools/isoreplacer)
- Test game with emulator [pcsx2](https://github.com/PCSX2/pcsx2)

## Screenshots
![](/screenshots/ATHN01A.png "Athens")
![](/screenshots/SEWR01.png "Sewerage")
