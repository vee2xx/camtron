electron-packager . camtron --platform=linux --arch=x64 --overwrite
electron-packager . camtron --platform=win32 --overwrite
electron-packager . camtron --platform=darwin --arch=x64 --overwrite

rm  -rf dist/camtron-darwin-x64/
rm  -rf dist/camtron/camtron-linux-x64/
rm  -rf dist/camtron/camtron-win32-x64/

mv camtron-darwin-x64/ dist/
mv camtron-linux-x64/ dist/
mv camtron-win32-x64/ dist/
