# Security Policy

## Reporting a Vulnerability

If you find a vulnerability in sastsweep and would like to report it (thank you 🙏), please DM me on Twitter [@_chebuya](https://x.com/_chebuya).

The main security concern I have with this tool are malicious repositories.

Since `sastsweep` is designed to be fed in a bunch of repos and scan them, this involves downloading the .zip file containing all the code and unzipping it on the system running `sastsweep`.

I have taken some measures to prevent people from using zip slip, a zip bomb, or just generally writing files where they should not be written - however, I doubt this implementation is perfect:

The unzipping routine is located here:

  https://github.com/chebuya/sastsweep/blob/main/common/util.go#L57-L129 
  
  https://github.com/chebuya/sastsweep/blob/main/common/util.go#27-L55
