# E-Voting
Bachelor's Thesis - Implementation of an e-voting system based on blockchain technology 

## Architecture

The system consists of the following components:
  - User Interface (included in a separate repository) - enables the end users to submit new votes and verify the existing ones. 
  - Blockchain - all blockchain nodes.
  - Blockchain Connector - a broker which enables communication between the web application and the blockchain.
  - Node Registry - contains information about the blockchain network.

![System Overview](https://user-images.githubusercontent.com/44197493/150641994-5b3a1d16-4092-40c0-8dce-287f2cec379c.png)

## Practical Byzantine Fault Tolerance (PBFT)

The running system is using the PBFT consensus mechanism. It works according to the following scheme:

![pbft](https://user-images.githubusercontent.com/44197493/150642145-c470cbd3-b38e-468f-8fb2-5df24755c774.png)

## Implementation Overview

The key logic revolves around the following classes (i.e. Golang structs):

![class diagrams (1)](https://user-images.githubusercontent.com/44197493/150642223-50ebed56-68d7-4bfb-a7d3-026dd695da13.png)

## Vote Casting and Verification

The vote casting procedure uses the system in the following manner:

![vote casting upscaled](https://user-images.githubusercontent.com/44197493/150642282-62c9c46d-0c85-44b9-b99a-0ad3a9695704.png)

After a user successfully casts a vote, a UUID token is returned. The token is bound to the corresponding user's vote (to avoid disclosing information such as the relation between particular users and votes). The tokens allow users to verify if their vote was submitted correctly and whether it has not been altered after the submission. The verification procedure is presented graphically below: 

![vote verification upscaled (1)](https://user-images.githubusercontent.com/44197493/150642269-536f3842-7126-47d6-9fa5-0ffeb9c057b1.png)
