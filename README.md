[![Go Report Card](https://goreportcard.com/badge/github.com/mogaika/god_of_war_browser)](https://goreportcard.com/report/github.com/mogaika/god_of_war_browser)

![gow_browser_logo](https://user-images.githubusercontent.com/3680954/28489831-6ec1c660-6edd-11e7-9b08-7c79b20196d8.png)

Tool helps to investigate files formats of the game.
Based on version: *God of War (2005)  NTSC(USA)  PS2DVD-9*

## Instruction
- Install golang compiler ([official instruction](https://golang.org/doc/install))
- `go get -u github.com/mogaika/god_of_war_browser` (this command download code and dependencies)
- Go to directory with required [tool](https://github.com/mogaika/god_of_war_browser/tree/master/tools) and run `go build` to compile
- Do not forget read README
- If you encounter any problem or have a question - welcome to our friendly [gitter chat](https://gitter.im/god_of_war/)

## Instruction to god_of_war_browser
- Open console and launch binary with
  - ```-iso "Path_to_game_ISO_file"``` if you have iso file (detection of second layer implemented)
  - ```-tok "Path_to_directory_with_GODOFWAR.TOK_and_PART?.pak_files"``` if you have unpacked iso
- Then open http://127.0.0.1:8000/ in your browser (address can be changed via ```-i Listen_IP:PORT```)

## What if I want to mod game?
You can! But it is hard at this point :(
- You can change textures in browser! Open TXR_ resource and use upload form (png,jpg,gif support). But first time you upload larger texture, it takes a while (~1-5min, depends on hard drive and memory) to rearrange resources in pack file to create free space.
- For other files follow:
  - Download required wads using god_of_war_browser web interface
  - Use [wadunpack](https://github.com/mogaika/god_of_war_browser/tree/master/tools/wadunpack) to unpack wad where you want to make change
  - Add/remove/modify game files, do not forget to modify *wad_meta.txt*. Some info about file formats you can find in [sources](https://github.com/mogaika/god_of_war_browser/tree/master/pack/wad) of god_of_war_browser.
  - Archive files to wad using [wadpack](https://github.com/mogaika/god_of_war_browser/tree/master/tools/wadpack)
  - Pack wad files in part1.pak using [packer](https://github.com/mogaika/god_of_war_browser/tree/master/tools/packer)
  - For test replace original part1.pak and godofwar.toc in your iso using [isoreplacer](https://github.com/mogaika/god_of_war_browser/tree/master/tools/isoreplacer)
- Test game with emulator [pcsx2](https://github.com/PCSX2/pcsx2). Do not forget to shutdown god_of_war_browser before run game in emulator (only of using same iso)

## Screenshots
![Athens](https://user-images.githubusercontent.com/3680954/28489832-6ec6697c-6edd-11e7-8ead-ed37e3870b15.png)
![Sewerage](https://user-images.githubusercontent.com/3680954/28489833-6ecbfc5c-6edd-11e7-9d63-1ca0b060ddec.png)


