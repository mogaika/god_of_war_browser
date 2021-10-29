# Strings locations
- Text showed to player in-game located in R_PERM.WAD => msgs_en.txt
- Text for in-game ui (upgrade menu/pause menu/settings) located at R_PERM.WAD => FLP_HUD
- Text for main menu located at R_SHELL.WAD => FLP_Shell.

# How to change text in msgs_en.txt
1. Open R_PERM.WAD and use download/upload buttons next to msgs_en.txt label
2. If you use custom fonts with changed encoding in flp, then make sure that this txt file saved in your codepage

# How to change static labels in FLP_ with custom charset
1. Open [BMFont](http://www.angelcode.com/products/bmfont/), select your font characters. I recommend to use this settings for export:
  - Export options:
    - Padding: 1
    - Spacing: 1
    - Texture width/height: 512/256
    - Bit depth: 32
    - Preset: White text with alpha
    - **File format: Binary**
    - **Textures: png**
  - Font settings:
    - **Charset: Unicode**
    - Size (px): 60
2. Save font with `Save bitmap font as..`
3. Open 'font_aliases.cfg' file in root of *god_of_war_browser* folder and fill it correspond to your lang coding page. For example in root of folder exists *font_aliases.ru.cp1251.cfg* file for cp1251 codepage. This is needed for correct live text editing in browser since for static labels game not storing strings, but render commands. Make sure that file saved in UTF-8 encoding.
4. Add *.fnt and *.png files in zip archive
5. Start *god_of_war_browser*, go to flp file that you want localize, choose font viewer on top of page and press "Import glyphs from BMFont file", choose zip file in dialog. Files in pack may be reordered first time you edit something in game package, this can take 3-15 mins.
6. Reload page and go to flp=>'Labels editor'. You can change font scale, blend color, x/y offsets and text of labels. You can preview changes and compare them side by side with original labels

# How to change strings in FLP_ scripts
1. If you can't find your text in labels section then check entire flp file dump using dump tab and "Download as json".
2. Change strings in scripts in json and upload it using "Upload from json" button. You should provide `--encoding` argument for gow browser if you imported font with custom encoding in previous section. Json file should be in UTF-8 encoding, provided encoding argument will be used on flp reading/writing stage. To list available encodings use `god_of_war_browser -listencodings`.
