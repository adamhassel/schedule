# Schedule

A (so far WIP) project that aims to generate a run schedule for something that needs to run occationally, while taking
power cost into account. The main motivation for making this is a pool circulation pump being controlled by a Shelly, so
the library will prioritize running in the daytime to optimize filtration during sunlight hours.

## Features/Wishlist

* Will read power prices, eventually online, but maybe from somewhere else as well? Who knows?
* Allows to set a maximum number of allowed hours to run during dark hours (geolocated from IP)
* Intended to be used as a library from some software which will handle the Shelly communication, so that part is not in
  scope here.
* So, yeah: Based on a list of power prices for a given hour in the day, generate a cost-optimized runnig schedule.
  Granularity is one hour forthis reason.