# Tetris Client
Websocket terminal tetris client. 

## Guide

### Controls
Controls are realised by buttons:
- a = move piece left
- s = move piece down
- d = move piece right
- q = rotate piece left
- e = rotate piece right

### Launch
Show help menu
```
./tetrisClient help
```

Launch menu
```
./tetrisClient
```

Create new session:
```
./tetrisClient create
```

Connect to existing session:
```
./tetrisClient connect 123456789
```

Show list of existing sessions:
```
./tetrisClient list
```

## Demo
The Websocket server runs on a remote computer and is controlled via an ssh connection on the left screen. Two websocket clients run on the same computer and connect to the same server

### Scoring tetris:
![Tetris.gif](./gifs/Tetris.gif)

### Full Game:
![DrawGame.gif](./gifs/DrawGame.gif)
