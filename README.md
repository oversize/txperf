# txperf creates transaction

... with a timestamp so that the time it took between tx submission
and tx adoption can be measured.

## Usage

```
# My keys stored in ~/.txperf/keys/
txperf key new
txperf key list --key KEYNAME       # With keyname just that
txperf key list                     # Without list all


# Addresses for keys
txperf address list --key KEYNAME   # With keyname just for that key
txperf address list                 # Without keyname list all (incl. key)


# Profiles combine keys and "txettings" like network
# json files in ~/.txperf/profiles/
txperf new profile
txperf get profiles

txperf run --key                    # In its first variant just the most basic one

```
