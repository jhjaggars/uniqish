# Uniqish

A log filtering program inspired by [uno](https://unomaly.com/blog/its-in-the-anomalies/).

After reading the above blog post, I set out to build something that _sounds_
similar.  To be clear, I *don't* know how `uno` works other than what is
described in the post, and I'm pretty sure it's more sophisticated than
described.

# How to use it

```
uniqish < /path/to/logfile
```

# How it works

By default the program will attempt to filter out log lines that look similar
to some defined amount of previous lines.  This is implemented by using a
frecency cache (the most recent and frequently accessed lines).

Each line is compared to each line in the cache by skipping a _common-ish_
prefix and then taking the edit distance.
