# Tasks

## TODO

- when gmenu is operated without calling quit immediately sometimes it hangs it's because key handlers and selections made doesn't hide and reset immediately perhaps atomicaly
- make the path where quit is called vs when it's not called and repeatedly used and reset closer and share more logic. eg calling hide and reset should happen on end for both
