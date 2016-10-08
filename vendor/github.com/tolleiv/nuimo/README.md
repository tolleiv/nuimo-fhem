# Golang library to interact with Nuimo devices

This library uses the [currantlabs/ble](https://github.com/currantlabs/ble) library and implements an interaction layer for [Senic Nuimo devices](https://www.senic.com/). Similar to [nathankunicki/nuimojs](https://github.com/nathankunicki/nuimojs) for NodeJS, it was a good inspiration for the library. More production-ready libraries can be find in the official [Senic](https://github.com/getsenic) repos.
 
## Disclaimer
 
At the moment this is a evenings project for me to learn Golang programming and built up some smart home / media control center know how. Feel free to suggest changes which change code and interaction to be more #Golang style.

## Example usage*

Please refer to the [currantlabs/ble](https://github.com/currantlabs/ble) documentation for the basic platform setup. Once the platform is ready run:

    go get github.com/tolleiv/nuimo
    # Check out the inputs:
    sudo go run src/github.com/tolleiv/nuimo/examples/inputs/main.go
    # Use the display
    sudo go run src/github.com/tolleiv/nuimo/examples/display/main.go

*this has been tested successfully on Linux only

## License 
 
 MIT License
