# Nuimo <-> FHEM connector 

This translates the commands from a [Nuimo](http://www.senic.com) smart controll into commands for the popular [FHEM](http://fhem.org) house automation system.
 
The mapping is done within the `scenes.yml`. This way Nuimo can basically control everything FHEM controls. For me that's some "smart" lights and some media equipment.

## Disclaimer
 
At the moment this is a evenings project to play with house automation and #golang. Feel free to suggest changes which change code and interaction to be more #Golang style.

## Example usage*

    go get github.com/tolleiv/nuimo-fhem
    # Copy and adjust the settings:
    cp $GOPATH/src/github.com/tolleiv/nuimo-fhem/scenes.yaml .
    # Check out the inputs:
    sudo go run $GOPATH/src/github.com/tolleiv/nuimo-fhem/main.go

*this has been tested successfully on Linux (RPi )only

## License 
 
 MIT License
