/*
  Information backup and recovery using the ideas presented in 
    https://www.notion.so/iden3/Identity-Backup-Retrieval-and-Restore-5be29311561e47c08710fdd03ab8b7b8 that can
    be used as a feature in Iden3 Identity Wallet

  In this example we generate an encrypted backup file with the different data structures that form the identity.
   In this exmaple, we are using dummy version of Claims. Merkle Tree, ZKP ... to show that we can encode and decode
   arbitrary data structures. The backup file contains a header that indicates the Key derivation algorithm
   employed, the parameters and the encryption protocol used. The libraries used can be expanded to support
   additional mechanisms (So far, for Key derivation only PBKDF2 with either SHA1 or SHA256 is supported, and only GCM 
   or no encryption are supported as encryption protocols).

  To recover the encryption key, we use Shamir's secret sharing algo (again, code can be extended to suport additional 
    protocols) to distribute the Key.

  For Finite Field Arithmetic, we use use a modified goff library version so that it provides a common interface
  to any type of field element used. In this example we use the BN256 prime 

*/

package main

import (
        "github.com/iden3/go-backup/ff"
     fc "github.com/iden3/go-backup/filecrypt"
        "fmt"
)

func main(){

   // Generates Main Key from BN256. This is our Identity operational Key. I am assuming
   // that the operational key is the one that enables us to regain the identity. 
   kOp := KeyOperational()
   
   // Retrieve Genesis ID -> Not really needed
   iD := NewID()

   // Generate Shares of our operational Key
   shares := GenerateShares(kOp)

   // Define Custodians -> Simulates the process of inviting a trusted entity to
   //   become our custodian, and sending N shares of the key upon acceptance of the invitation. In this example
   //   we generate a QR, but we could encrypt the share information and send it via some
   //   alternative mechanism (EMAIL, TELEGRAM, P2P connection). 
   //
   //   During the process of adding a custodian, we also could link this custodian with
   //   the contact details of this person available in the agenda so that we know 
   //   how to contact them in the future.

      // assign first share
   AddCustodian("Pedrito", QR ,shares,1,1 )
      // assign second share
   AddCustodian("Faustino", QR ,shares,2,1 )
      // assign third and fourth share
   AddCustodian("Sara Baras",QR  ,shares,3,2 )
      // assign 5th share
   AddCustodian("Sergio", QR ,shares,5,1 )
      // assign 6th share
   AddCustodian("Raul", QR ,shares,6,1 )

   // Define which information is included in Backup file. Contents of the backup are not important right
   // now. It is just to show how easy it is to build the backup file.
   // At this point, the buildup is static (I need to have a switch case in AddToBackup with all possible
   // data stucture types that can be added to the backup, but there should be a more dynamic way...)

     // Add Rx Claims
   AddToBackup(CLAIMS,        claims,     ENCRYPT)
     // Add wallet configuration
   AddToBackup(WALLET_CONFIG, wallet,     ENCRYPT)
     // Add generated ZKP 
   AddToBackup(ZKP_INFO,      zKPData,    ENCRYPT)
     // Add Merkle Tree -> If we haven't created any claims, we don't need to store
     // Merkle Tree because we could regenerate it
   AddToBackup(MERKLE_TREE,   merkleTree, ENCRYPT)
     // Add Custodian information (contact details) -> unencrypted
   AddToBackup(CUSTODIAN,     custodians, DONT_ENCRYPT)
     // Add Genesis ID) -> unencrypted
   AddToBackup(GENID,         iD,         DONT_ENCRYPT)
     // Add SSharing info. We need Prime number and protocol used (Shamir) -> unencrypted
   AddToBackup(SSHARING,       secret_cfg,     DONT_ENCRYPT)
     // Add Shares. We heed to keep a list of at least outstanding shares in case
     //  we want to redistribute in the future. in this example I keep all for simplicity.
   AddToBackup(SHARES,        shares,     ENCRYPT)

   // Generate Backupfile -> Here we select the Key derivation algo and the encryption mechanism used
   //  for encrypted sections. Also not, that we can mix encrypted and non-encrpyted information in the
   // same baclup file  
   CreateBackup(fc.FC_KEY_T_PBKDF2, fc.FC_HASH_SHA256, fc.FC_GCM, BACKUP_DIR+BACKUP_FILE, kOp)

   // We lost our phone.  We need to reinstall wallet in new phone and retrieve backup
   // from cloud services. 

   // Decode unencrypted backup file as it may contain some info 
   // (custodian contact info, genesis id, sharing)

   // During this fist stage, we only recover nonencrypted data as we still don't have the key.
   info := Decode(BACKUP_DIR+BACKUP_FILE, nil)

   // Retrieve custodian info -> I am assuming thattodian info contains a reminder of who my custodians were.
   //   In this example, I am using the custodian information to locate the QR code

   //   Retrieval converts generic array info to specific data type. Ideally, it should
   //   be  a single function and not a family of functions depending on type, but
   //   for now it was easier to do it like this
   retrieved_custodians := RetrieveCustodians(info)
   collected_shares := make([]map[uint64]ff.Element, 0)
 
   res := CheckEqual(custodians, retrieved_custodians)
   if res {
      fmt.Println("Retrieved Custodians .... OK" )
   } else {
      fmt.Println("Retrieved Custodians .... KO" )
   }

   // Retreive sharing info -> Finite Field information and protocol (Shamir's secret sharing) required to
   //    regenerate the KEY. It is unencrypted
   retrieved_sharing  := RetrieveSSharing(info)
   res = CheckEqual(&secret_cfg, retrieved_sharing)
   if res {
      fmt.Println("Retrieved Sharing Conf .... OK" )
   } else {
      fmt.Println("Retrieved Sharing Conf .... KO" )
   }

   // Retreive genesis ID -> Not used, but maybe in a real use case it is useful to have it available
   retrieved_iD       := RetrieveID(info)
   res = CheckEqual(iD, retrieved_iD)
   if res {
      fmt.Println("Retrieved Genesis ID .... OK" )
   } else {
      fmt.Println("Retrieved Genesis ID .... KO" )
   }

   // Contact custodians and retrieve shares
   // We simulate here that somehow we contact the custodians using the info in the backup 
   //    Out of the 5 custodians we had, we ony contacted   three.
   // The custodian then sends the share in P2P channel. In our case, we assume that we are 
   //  face to face and the custodian gfenerates a QR that we can scan. 

   
   for _, custodian := range(retrieved_custodians){
      collected_shares = append(collected_shares, ScanQRShare(&custodian)...)
   }

   // Generate Key 
   //   Using the collected shares, regenerate Key
   retrieved_kOp := GenerateKey(collected_shares, retrieved_sharing)
   res = CheckEqual(kOp, retrieved_kOp)
   if res {
      fmt.Println("Retrieved kOp .... OK" )
   } else {
      fmt.Println("Retrieved kOp .... KO" )
   }

   
   // Decode and Decrypt backup file -> With the generated kOp, try to decrypt file.
   //   kOp is not used directly. We use a Key Derivation Function. All parameters for this 
   //   function are public (except for the Key) and are in the encryption block header
   info = Decode(BACKUP_DIR+BACKUP_FILE, retrieved_kOp)
   
   // With the decrpyted and decoded information, retrieve all information we stored and check 
   // if it is equal than the original
   retrieved_claims := RetrieveClaims(info)
   res = CheckEqual(claims, retrieved_claims)
   if res {
      fmt.Println("Retrieved Claims .... OK" )
   } else {
      fmt.Println("Retrieved Claims .... KO" )
   }
   
   retrieved_wallet := RetrieveWallet(info)
   res = CheckEqual(wallet, retrieved_wallet)
   if res {
      fmt.Println("Retrieved Wallet .... OK" )
   } else {
      fmt.Println("Retrieved Wallet .... KO" )
   }
  
   retrieved_mt     := RetrieveMT(info)
   res = CheckEqual(merkleTree, retrieved_mt)
   if res {
      fmt.Println("Retrieved MT .... OK" )
   } else {
      fmt.Println("Retrieved MT .... KO" )
   }
   
   retrieved_shares := RetrieveShares(info)
   res = CheckEqual(shares, retrieved_shares)
   if res {
      fmt.Println("Retrieved Shares .... OK" )
   } else {
      fmt.Println("Retrieved Shares .... KO" )
   }
   
   retrieved_zkp    := RetrieveZKP(info)
   res = CheckEqual(zKPData, retrieved_zkp)
   if res {
      fmt.Println("Retrieved ZKP .... OK" )
   } else {
      fmt.Println("Retrieved ZKP .... KO" )
   }
   

   // Last step is to restore identity using retrieved kOp. Since we do not store any claims, 
   // we could regenerate the identiy usingt the key.
}
        