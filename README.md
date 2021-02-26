# 🤖 Go-Rosbot-Collector

Scrapes Ros-Bot's user activity page, and returns the parsed server updates.

## Table of Contents

- [Usage](#usage)
  - [New Client](#new-client)
  - [Parsing](#parsing)
    - [Defaults](#defaults)
    - [Custom](#custom)
  - [Errors](#errors)
  - [Types](#types)
    - [Server Update](#server-update)
    - [Legendary Item](#legendary-item)
- [Contributions](#contributions)
- [License](#license)

<hr/>

## Usage

### New Client

```go
type Client interface {
    // ParseWithDefaults returns a slice of Ros-Bot server updates based on the default parsing
    // configuration.
    ParseWithDefaults(ctx context.Context) ([]*ServerUpdate, error)
    // ParseWithConfig returns a slice of Ros-Bot server updates based on the provided
    // parsing configuration.
    ParseWithConfig(ctx context.Context, config *ParserConfig) ([]*ServerUpdate, error)
}
```

```go
rbc, err := rosbotcollector.NewClient("your-username", "password")
if err != nil {
	...
}
```

### Parsing

```go
type ParserConfig struct {
  Destinations []Destination
  RarityLevel  Rarity
  Quality      Quality 
  Page         int8
}
```

```go
c := rosbotcollector.NewParseConfig()

                OR

c := rosbotcollector.ParseConfig{
	...
}
```

#### Defaults

Uses the *de-facto* configuration.

```go
ParserConfig{
    Destinations: []Destination{},
    RarityLevel:  RarityNonAncient,
    Quality:      QualityAll,
    Page:         1,
}
```

```go
u, err := rbc.ParseWithDefaults(ctx)
if err != nil {
	...
}
```

#### Custom

Uses a custom configuration object.

```go
c := rosbotcollector.NewParseConfig()

                OR

c := rosbotcollector.ParseConfig{...}
```

```go
u, err := rbc.ParseWithConfig(ctx, &c)
if err != nil {
	...
}
```

### Errors

`ErrBadCredentials` is returned when the login attempt has failed.

`ErrNoFormBuildID` is returned when `form_build_id` could not be parsed from response body.

`ErrNoActivityEndpoint` is returned when the activity endpoint could not be parsed from response body.

`ErrCookiesRefresh` is returned when the attempt to refresh user cookies has failed.

## Types

### Server Update

Corresponds to a Ros-Bot server update.

```go
type ServerUpdate struct {
    Items           []*LegendaryItem `json:"legendaries"`
    ServerTimestamp time.Time        `json:"server_timestamp"`
}
```

### Legendary Item

Corresponds to an in-game item of "legendary" quality.

```go
type LegendaryItem struct {
  Name         string
  IsIdentified bool
  Quality      Quality 
  Rarity       Rarity
  Destination  Destination
  Stats        string
}
```

#### Quality

```go
QualityAll    Quality = "*"
QualityNormal Quality = "NORMAL"
QualitySet    Quality = "SET"
```

#### Rarity 

```go
RarityPrimal     Rarity = "PRIMAL"
RarityAncient    Rarity = "ANCIENT"
RarityNonAncient Rarity = "NON-ANCIENT"
```

#### Destination 

```go
DestinationStashed  Destination = "STASHED"
DestinationSalvaged Destination = "SALVAGED"
DestinationSold     Destination = "SOLD"
```

## Contributions 

- [ ] Improve item property parsing (=? weapon, armour, ring, etc).

- [ ] Improve item stat. parsing (base, primary, secondary, power, sockets)

- [ ] Expand parsing to include items of all qualities. 

## License

[Apache License](LICENSE)

Maximilien Zaleski © 2019
