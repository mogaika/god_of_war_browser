[![Go Report Card](https://goreportcard.com/badge/github.com/mogaika/god_of_war_browser)](https://goreportcard.com/report/github.com/mogaika/god_of_war_browser)
[![Build status](https://ci.appveyor.com/api/projects/status/n4w8rkn30sl6oqbp/branch/master?svg=true)](https://ci.appveyor.com/project/mogaika/god-of-war-browser/branch/master)

![gow_browser_logo](https://user-images.githubusercontent.com/3680954/28489831-6ec1c660-6edd-11e7-9b08-7c79b20196d8.png)

Tool that helps to investigate files formats of the game.
Some functional broken on *Jp version*.

### [Download lastet build](https://ci.appveyor.com/project/mogaika/god-of-war-browser/branch/master/artifacts)

## Instruction
- Download and unzip [lastet build](https://ci.appveyor.com/project/mogaika/god-of-war-browser/branch/master/artifacts)
- Open console and launch binary with parameters:
  - Archive source
    - ```-iso "Path_to_game_ISO_file"``` if you have iso file. Detection of second layer implemented (it is not supported by almost every virtual drive software)
    - ```-toc "Path_to_directory_with_GODOFWAR.TOC_and_PART?.pak_files"``` if you have pak and toc files
    - ```-dir "Path_to_directory_with_WAD_files"``` if you have wad files
    - ```-psarc "Path_to_psarc_file"``` if you have psarc archive
  - Playstation version using one of the
    ```-ps ps2``` ```-ps ps3``` ```-ps psvita```
  - Target game
    ```-gowversion 1``` for GoW I or ```-gowversion 2``` for GoW II
- Open http://127.0.0.1:8000/ in your browser (address can be changed via ```-i Listen_IP:PORT```)

## What if I want to mod game?
You can! But it is hard at this time :(
- Remember! First time you upload larger file, it takes a while (~1-5min, depends on hard drive) to rearrange resources in pack file to create free space (Check console log for progress).
- Also remember that the tool is not ideal, and I ask to make backups of the original iso and of your progress.
- You can download resources, change them in hex editor and upload back using browser.
- You can reupload textures right in browser! Open TXR_ resource and use upload form (png,jpg,gif support).
- You can change UI labels inside FLP_ resources. And even create new fonts (FLP related stuff may be broken buld to build)
- Legacy flow of modifications:
  - Download required wads using god_of_war_browser web interface
  - Use [wadunpack](https://github.com/mogaika/god_of_war_browser/tree/master/tools/wadunpack) to unpack wad where you want to make change
  - Add/remove/modify game files, do not forget to modify *wad_meta.txt*. Some info about file formats you can find in [sources](https://github.com/mogaika/god_of_war_browser/tree/master/pack/wad) of god_of_war_browser.
  - Archive files to wad using [wadpack](https://github.com/mogaika/god_of_war_browser/tree/master/tools/wadpack)
  - Pack wad files in part1.pak using [packer](https://github.com/mogaika/god_of_war_browser/tree/master/tools/packer)
  - For test replace original part1.pak and godofwar.toc in your iso using [isoreplacer](https://github.com/mogaika/god_of_war_browser/tree/master/tools/isoreplacer)
- Test game with emulator [pcsx2](https://github.com/PCSX2/pcsx2). Do not forget to shutdown god_of_war_browser before run game in emulator (only if using same iso file)

## What about animations?
Just open any obj file. Supported parsing of:
- Joint rotation
- Joint position
- Texture UV
- Texture flipbook

![R_SKC](https://user-images.githubusercontent.com/3680954/71230603-bd897e00-2303-11ea-8f5e-ef84d81dfef0.gif)

## Screenshots
![Athens](https://user-images.githubusercontent.com/3680954/28489832-6ec6697c-6edd-11e7-8ead-ed37e3870b15.png)
![Sewerage](https://user-images.githubusercontent.com/3680954/28489833-6ecbfc5c-6edd-11e7-9d63-1ca0b060ddec.png)
