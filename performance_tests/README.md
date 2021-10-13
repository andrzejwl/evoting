# Test Suite for Performance Comparison between PoW and PBFT

-----

## Throughput

Goal: find out which algorithm accepts new transactions faster.

Tests may be concluded using the default `docker-compose` environment but real-life scenario network conditions (latency, packet loss) may be simulated using [traffic control](https://man7.org/linux/man-pages/man8/tc.8.html).

-----

## Power Usage

Goal: find out which consensus mechanism is more power consuming and demands more CPU resources.

-----

## Network Usage

Goal: find out which consensus mechanism has the highest communication overhead.
