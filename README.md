[![Go Report Card](https://goreportcard.com/badge/github.com/mogaika/god_of_war_browser)](https://goreportcard.com/report/github.com/mogaika/god_of_war_browser)
[![Build status](https://ci.appveyor.com/api/projects/status/n4w8rkn30sl6oqbp/branch/master?svg=true)](https://ci.appveyor.com/project/mogaika/god-of-war-browser/branch/master)

![gow_browser_logo](https://user-images.githubusercontent.com/3680954/28489831-6ec1c660-6edd-11e7-9b08-7c79b20196d8.png)

A tool that allows browsing and investigating file formats of the game.
Some functions are broken on the *Japanese version of the game (NTSC-J)*.

### [Download latest build](https://ci.appveyor.com/project/mogaika/god-of-war-browser/branch/master/artifacts)

### [Join Discord](https://discord.gg/u6x3Z9v6Ed)

## Instructions
- Download and unzip [latest build](https://ci.appveyor.com/project/mogaika/god-of-war-browser/branch/master/artifacts)
- Open a console window and launch the binary with chosen parameters:
  - Archive source
    - ```-iso "Path_to_ISO_file"``` if you have an .iso file. Detection of second layer implemented (it is not supported by almost every virtual drive software)
    - ```-toc "Path_to_directory_with_GODOFWAR.TOC_and_PART?.PAK_files"``` if you have .PAK and .TOC files
    - ```-dir "Path_to_directory_with_WAD_files"``` if you have .WAD files
    - ```-psarc "Path_to_psarc_file"``` if you have a psarc archive
  - Chosen playstation version
    ```-ps ps2``` ```-ps ps3``` ```-ps psvita```
  - Target game
    ```-gowversion 1``` for GoW I or ```-gowversion 2``` for GoW II
- Open http://127.0.0.1:8000/ in your browser (address can be changed via ```-i Listen_IP:PORT```)
- In 3d view you can use:
	- `LMB` to look around. While holding `LMB` use `W` or `S` to move target forwards or backwards
	- `Shift + LMB` or `MMB` to move target in horizontal plane
	- Mouse wheel to zoom to/from target

## What if I want to mod the game?
You can! But it is hard at this time :(
- Remember: First time you upload a larger file, it takes a while (~1-5min, depending on your hard drive) to rearrange resources in pack file to create free space (You can check the console log for progress).
- Also remember that the tool is not perfect, and you should make backups of the original .iso and of your progress.
- You can download resources, change them in a hex editor and upload them back using the browser UI.
- You can reupload textures right in the browser window! Open any TXR_ resource and use the upload form (png,jpg,gif support).
- You can change UI labels inside FLP_ resources, and even create new fonts! (FLP related stuff may be broken from build to build)
- Legacy flow of modifications:
  - Download required .WADs using the god_of_war_browser web interface
  - Use [wadunpack](https://github.com/mogaika/god_of_war_browser/tree/master/tools/wadunpack) to unpack the .WAD file where you want to make changes
  - Add/remove/modify the game files, but do not forget to modify *wad_meta.txt*. Some info about the file formats can be found in the [sources](https://github.com/mogaika/god_of_war_browser/tree/master/pack/wad) of god_of_war_browser.
  - Archive the files back into .WAD using [wadpack](https://github.com/mogaika/god_of_war_browser/tree/master/tools/wadpack)
  - Pack wad files in part1.pak using [packer](https://github.com/mogaika/god_of_war_browser/tree/master/tools/packer)
  - For testing replace the original PART1.PAK and GODOFWAR.TOC in your .iso using [isoreplacer](https://github.com/mogaika/god_of_war_browser/tree/master/tools/isoreplacer)
- Test the game using an emulator ([pcsx2](https://github.com/PCSX2/pcsx2)). Do not forget to close and exit god_of_war_browser before running the game in the emulator! (This applies only if you will be using the same .iso file)

## What about animations?
Just open any obj file! Supported parsing of:
- Joint rotation
- Joint position
- Texture UV
- Texture flipbook

![R_SKC](https://user-images.githubusercontent.com/3680954/71230603-bd897e00-2303-11ea-8f5e-ef84d81dfef0.gif)

## Screenshots
![Athens](https://user-images.githubusercontent.com/3680954/28489832-6ec6697c-6edd-11e7-8ead-ed37e3870b15.png)
![Sewerage](https://user-images.githubusercontent.com/3680954/28489833-6ecbfc5c-6edd-11e7-9d63-1ca0b060ddec.png)
