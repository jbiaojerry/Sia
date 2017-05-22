package main

import (
	"fmt"
	"math/big"

	"github.com/bgentry/speakeasy"
	"github.com/spf13/cobra"

	"github.com/NebulousLabs/Sia/api"
	"github.com/NebulousLabs/Sia/types"
)

var (
	walletCmd = &cobra.Command{
		Use:   "wallet",
		Short: "Perform wallet actions",
		Long: `Generate a new address, send coins to another wallet, or view info about the wallet.

Units:
The smallest unit of siacoins is the hasting. One siacoin is 10^24 hastings. Other supported units are:
  pS (pico,  10^-12 SC)
  nS (nano,  10^-9 SC)
  uS (micro, 10^-6 SC)
  mS (milli, 10^-3 SC)
  SC
  KS (kilo, 10^3 SC)
  MS (mega, 10^6 SC)
  GS (giga, 10^9 SC)
  TS (tera, 10^12 SC)`,
		Run: wrap(walletbalancecmd),
	}

	walletAddressCmd = &cobra.Command{
		Use:   "address",
		Short: "Get a new wallet address",
		Long:  "Generate a new wallet address from the wallet's primary seed.",
		Run:   wrap(walletaddresscmd),
	}

	walletAddressesCmd = &cobra.Command{
		Use:   "addresses",
		Short: "List all addresses",
		Long:  "List all addresses that have been generated by the wallet.",
		Run:   wrap(walletaddressescmd),
	}

	walletChangepasswordCmd = &cobra.Command{
		Use:   "change-password",
		Short: "Change the wallet password",
		Long:  "Change the encryption password of the wallet, re-encrypting all keys + seeds kept by the wallet.",
		Run:   wrap(walletchangepasswordcmd),
	}

	walletInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize and encrypt a new wallet",
		Long: `Generate a new wallet from a randomly generated seed, and encrypt it.
By default the wallet encryption / unlock password is the same as the generated seed.`,
		Run: wrap(walletinitcmd),
	}

	walletInitSeedCmd = &cobra.Command{
		Use:   "init-seed",
		Short: "Initialize and encrypt a new wallet using a pre-existing seed",
		Long:  `Initialize and encrypt a new wallet using a pre-existing seed.`,
		Run:   wrap(walletinitseedcmd),
	}

	walletLoadCmd = &cobra.Command{
		Use:   "load",
		Short: "Load a wallet seed, v0.3.3.x wallet, or siag keyset",
		// Run field is not set, as the load command itself is not a valid command.
		// A subcommand must be provided.
	}

	walletLoad033xCmd = &cobra.Command{
		Use:   "033x [filepath]",
		Short: "Load a v0.3.3.x wallet",
		Long:  "Load a v0.3.3.x wallet into the current wallet",
		Run:   wrap(walletload033xcmd),
	}

	walletLoadSeedCmd = &cobra.Command{
		Use:   `seed`,
		Short: "Add a seed to the wallet",
		Long:  "Loads an auxiliary seed into the wallet.",
		Run:   wrap(walletloadseedcmd),
	}

	walletLoadSiagCmd = &cobra.Command{
		Use:     `siag [filepath,...]`,
		Short:   "Load siag key(s) into the wallet",
		Long:    "Load siag key(s) into the wallet - typically used for siafunds.",
		Example: "siac wallet load siag key1.siakey,key2.siakey",
		Run:     wrap(walletloadsiagcmd),
	}

	walletLockCmd = &cobra.Command{
		Use:   "lock",
		Short: "Lock the wallet",
		Long:  "Lock the wallet, preventing further use",
		Run:   wrap(walletlockcmd),
	}

	walletSeedsCmd = &cobra.Command{
		Use:   "seeds",
		Short: "View information about your seeds",
		Long:  "View your primary and auxiliary wallet seeds.",
		Run:   wrap(walletseedscmd),
	}

	walletSendCmd = &cobra.Command{
		Use:   "send",
		Short: "Send either siacoins or siafunds to an address",
		Long:  "Send either siacoins or siafunds to an address",
		// Run field is not set, as the send command itself is not a valid command.
		// A subcommand must be provided.
	}

	walletSendSiacoinsCmd = &cobra.Command{
		Use:   "siacoins [amount] [dest]",
		Short: "Send siacoins to an address",
		Long: `Send siacoins to an address. 'dest' must be a 76-byte hexadecimal address.
'amount' can be specified in units, e.g. 1.23KS. Run 'wallet --help' for a list of units.
If no unit is supplied, hastings will be assumed.

A miner fee of 10 SC is levied on all transactions.`,
		Run: wrap(walletsendsiacoinscmd),
	}

	walletSendSiafundsCmd = &cobra.Command{
		Use:   "siafunds [amount] [dest]",
		Short: "Send siafunds",
		Long: `Send siafunds to an address, and transfer the claim siacoins to your wallet.
Run 'wallet send --help' to see a list of available units.`,
		Run: wrap(walletsendsiafundscmd),
	}

	walletSweepCmd = &cobra.Command{
		Use:   "sweep",
		Short: "Sweep siacoins and siafunds from a seed.",
		Long: `Sweep siacoins and siafunds from a seed. The outputs belonging to the seed
will be sent to your wallet.`,
		Run: wrap(walletsweepcmd),
	}

	walletBalanceCmd = &cobra.Command{
		Use:   "balance",
		Short: "View wallet balance",
		Long:  "View wallet balance, including confirmed and unconfirmed siacoins and siafunds.",
		Run:   wrap(walletbalancecmd),
	}

	walletTransactionsCmd = &cobra.Command{
		Use:   "transactions",
		Short: "View transactions",
		Long:  "View transactions related to addresses spendable by the wallet, providing a net flow of siacoins and siafunds for each transaction",
		Run:   wrap(wallettransactionscmd),
	}

	walletUnlockCmd = &cobra.Command{
		Use:   `unlock`,
		Short: "Unlock the wallet",
		Long:  "Decrypt and load the wallet into memory",
		Run:   wrap(walletunlockcmd),
	}
)

const askPasswordText = "We need to encrypt the new data using the current wallet password, please provide: "

const newPasswordText = "Please provide the current password of the wallet: "
const oldPasswordText = "Please provide the new password for the wallet: "

// walletaddresscmd fetches a new address from the wallet that will be able to
// receive coins.
func walletaddresscmd() {
	addr := new(api.WalletAddressGET)
	err := getAPI("/wallet/address", addr)
	if err != nil {
		die("Could not generate new address:", err)
	}
	fmt.Printf("Created new address: %s\n", addr.Address)
}

// walletaddressescmd fetches the list of addresses that the wallet knows.
func walletaddressescmd() {
	addrs := new(api.WalletAddressesGET)
	err := getAPI("/wallet/addresses", addrs)
	if err != nil {
		die("Failed to fetch addresses:", err)
	}
	for _, addr := range addrs.Addresses {
		fmt.Println(addr)
	}
}

// walletchangepasswordcmd changes the password of the wallet.
func walletchangepasswordcmd() {
	oldPassword, err := speakeasy.Ask(oldPasswordText)
	if err != nil {
		die("Reading password failed:", err)
	}
	newPassword, err := speakeasy.Ask(newPasswordText)
	if err != nil {
		die("Reading password failed:", err)
	}
	qs := fmt.Sprintf("encryptionpassword=%s&newpassword=%s", oldPassword, newPassword)
	err = post("/wallet/changepassword", qs)
	if err != nil {
		die("Changing the password failed:", err)
	}
	fmt.Println("Password changed sucessfully.")
}

// walletinitcmd encrypts the wallet with the given password
func walletinitcmd() {
	var er api.WalletInitPOST
	qs := fmt.Sprintf("dictionary=%s", "english")
	if initPassword {
		password, err := speakeasy.Ask("Wallet password: ")
		if err != nil {
			die("Reading password failed:", err)
		}
		qs += fmt.Sprintf("&encryptionpassword=%s", password)
	}
	if initForce {
		qs += "&force=true"
	}
	err := postResp("/wallet/init", qs, &er)
	if err != nil {
		die("Error when encrypting wallet:", err)
	}
	fmt.Printf("Recovery seed:\n%s\n\n", er.PrimarySeed)
	if initPassword {
		fmt.Printf("Wallet encrypted with given password\n")
	} else {
		fmt.Printf("Wallet encrypted with password:\n%s\n", er.PrimarySeed)
	}
}

// walletinitseedcmd initializes the wallet from a preexisting seed.
func walletinitseedcmd() {
	seed, err := speakeasy.Ask("Seed: ")
	if err != nil {
		die("Reading seed failed:", err)
	}
	qs := fmt.Sprintf("&seed=%s&dictionary=%s", seed, "english")
	if initPassword {
		password, err := speakeasy.Ask("Wallet password: ")
		if err != nil {
			die("Reading password failed:", err)
		}
		qs += fmt.Sprintf("&encryptionpassword=%s", password)
	}
	if initForce {
		qs += "&force=true"
	}
	err = post("/wallet/init/seed", qs)
	if err != nil {
		die("Could not initialize wallet from seed:", err)
	}
	if initPassword {
		fmt.Println("Wallet initialized and encrypted with given password.")
	} else {
		fmt.Println("Wallet initialized and encrypted with seed.")
	}
}

// walletload033xcmd loads a v0.3.3.x wallet into the current wallet.
func walletload033xcmd(source string) {
	password, err := speakeasy.Ask(askPasswordText)
	if err != nil {
		die("Reading password failed:", err)
	}
	qs := fmt.Sprintf("source=%s&encryptionpassword=%s", abs(source), password)
	err = post("/wallet/033x", qs)
	if err != nil {
		die("Loading wallet failed:", err)
	}
	fmt.Println("Wallet loading successful.")
}

// walletloadseedcmd adds a seed to the wallet's list of seeds
func walletloadseedcmd() {
	seed, err := speakeasy.Ask("New seed: ")
	if err != nil {
		die("Reading seed failed:", err)
	}
	password, err := speakeasy.Ask(askPasswordText)
	if err != nil {
		die("Reading password failed:", err)
	}
	qs := fmt.Sprintf("encryptionpassword=%s&seed=%s&dictionary=%s", password, seed, "english")
	err = post("/wallet/seed", qs)
	if err != nil {
		die("Could not add seed:", err)
	}
	fmt.Println("Added Key")
}

// walletloadsiagcmd loads a siag key set into the wallet.
func walletloadsiagcmd(keyfiles string) {
	password, err := speakeasy.Ask(askPasswordText)
	if err != nil {
		die("Reading password failed:", err)
	}
	qs := fmt.Sprintf("keyfiles=%s&encryptionpassword=%s", keyfiles, password)
	err = post("/wallet/siagkey", qs)
	if err != nil {
		die("Loading siag key failed:", err)
	}
	fmt.Println("Wallet loading successful.")
}

// walletlockcmd locks the wallet
func walletlockcmd() {
	err := post("/wallet/lock", "")
	if err != nil {
		die("Could not lock wallet:", err)
	}
}

// walletseedcmd returns the current seed {
func walletseedscmd() {
	var seedInfo api.WalletSeedsGET
	err := getAPI("/wallet/seeds", &seedInfo)
	if err != nil {
		die("Error retrieving the current seed:", err)
	}
	fmt.Println("Primary Seed:")
	fmt.Println(seedInfo.PrimarySeed)
	if len(seedInfo.AllSeeds) == 1 {
		// AllSeeds includes the primary seed
		return
	}
	fmt.Println()
	fmt.Println("Auxiliary Seeds:")
	for _, seed := range seedInfo.AllSeeds {
		if seed == seedInfo.PrimarySeed {
			continue
		}
		fmt.Println() // extra newline for readability
		fmt.Println(seed)
	}
}

// walletsendsiacoinscmd sends siacoins to a destination address.
func walletsendsiacoinscmd(amount, dest string) {
	hastings, err := parseCurrency(amount)
	if err != nil {
		die("Could not parse amount:", err)
	}
	err = post("/wallet/siacoins", fmt.Sprintf("amount=%s&destination=%s", hastings, dest))
	if err != nil {
		die("Could not send siacoins:", err)
	}
	fmt.Printf("Sent %s hastings to %s\n", hastings, dest)
}

// walletsendsiafundscmd sends siafunds to a destination address.
func walletsendsiafundscmd(amount, dest string) {
	err := post("/wallet/siafunds", fmt.Sprintf("amount=%s&destination=%s", amount, dest))
	if err != nil {
		die("Could not send siafunds:", err)
	}
	fmt.Printf("Sent %s siafunds to %s\n", amount, dest)
}

// walletbalancecmd retrieves and displays information about the wallet.
func walletbalancecmd() {
	status := new(api.WalletGET)
	err := getAPI("/wallet", status)
	if err != nil {
		die("Could not get wallet status:", err)
	}
	encStatus := "Unencrypted"
	if status.Encrypted {
		encStatus = "Encrypted"
	}
	if !status.Unlocked {
		fmt.Printf(`Wallet status:
%v, Locked
Unlock the wallet to view balance
`, encStatus)
		return
	}

	unconfirmedBalance := status.ConfirmedSiacoinBalance.Add(status.UnconfirmedIncomingSiacoins).Sub(status.UnconfirmedOutgoingSiacoins)
	var delta string
	if unconfirmedBalance.Cmp(status.ConfirmedSiacoinBalance) >= 0 {
		delta = "+" + currencyUnits(unconfirmedBalance.Sub(status.ConfirmedSiacoinBalance))
	} else {
		delta = "-" + currencyUnits(status.ConfirmedSiacoinBalance.Sub(unconfirmedBalance))
	}

	fmt.Printf(`Wallet status:
%s, Unlocked
Confirmed Balance:   %v
Unconfirmed Delta:  %v
Exact:               %v H
Siafunds:            %v SF
Siafund Claims:      %v H
`, encStatus, currencyUnits(status.ConfirmedSiacoinBalance), delta,
		status.ConfirmedSiacoinBalance, status.SiafundBalance, status.SiacoinClaimBalance)
}

// walletsweepcmd sweeps coins and funds from a seed.
func walletsweepcmd() {
	seed, err := speakeasy.Ask("Seed: ")
	if err != nil {
		die("Reading seed failed:", err)
	}

	var swept api.WalletSweepPOST
	err = postResp("/wallet/sweep/seed", fmt.Sprintf("seed=%s&dictionary=%s", seed, "english"), &swept)
	if err != nil {
		die("Could not sweep seed:", err)
	}
	fmt.Printf("Swept %v and %v SF from seed.\n", currencyUnits(swept.Coins), swept.Funds)
}

// wallettransactionscmd lists all of the transactions related to the wallet,
// providing a net flow of siacoins and siafunds for each.
func wallettransactionscmd() {
	wtg := new(api.WalletTransactionsGET)
	err := getAPI("/wallet/transactions?startheight=0&endheight=10000000", wtg)
	if err != nil {
		die("Could not fetch transaction history:", err)
	}

	fmt.Println("    [height]                                                   [transaction id]    [net siacoins]   [net siafunds]")
	txns := append(wtg.ConfirmedTransactions, wtg.UnconfirmedTransactions...)
	for _, txn := range txns {
		// Determine the number of outgoing siacoins and siafunds.
		var outgoingSiacoins types.Currency
		var outgoingSiafunds types.Currency
		for _, input := range txn.Inputs {
			if input.FundType == types.SpecifierSiacoinInput && input.WalletAddress {
				outgoingSiacoins = outgoingSiacoins.Add(input.Value)
			}
			if input.FundType == types.SpecifierSiafundInput && input.WalletAddress {
				outgoingSiafunds = outgoingSiafunds.Add(input.Value)
			}
		}

		// Determine the number of incoming siacoins and siafunds.
		var incomingSiacoins types.Currency
		var incomingSiafunds types.Currency
		for _, output := range txn.Outputs {
			if output.FundType == types.SpecifierMinerPayout {
				incomingSiacoins = incomingSiacoins.Add(output.Value)
			}
			if output.FundType == types.SpecifierSiacoinOutput && output.WalletAddress {
				incomingSiacoins = incomingSiacoins.Add(output.Value)
			}
			if output.FundType == types.SpecifierSiafundOutput && output.WalletAddress {
				incomingSiafunds = incomingSiafunds.Add(output.Value)
			}
		}

		// Convert the siacoins to a float.
		incomingSiacoinsFloat, _ := new(big.Rat).SetFrac(incomingSiacoins.Big(), types.SiacoinPrecision.Big()).Float64()
		outgoingSiacoinsFloat, _ := new(big.Rat).SetFrac(outgoingSiacoins.Big(), types.SiacoinPrecision.Big()).Float64()

		// Print the results.
		if txn.ConfirmationHeight < 1e9 {
			fmt.Printf("%12v", txn.ConfirmationHeight)
		} else {
			fmt.Printf(" unconfirmed")
		}
		fmt.Printf("%67v%15.2f SC", txn.TransactionID, incomingSiacoinsFloat-outgoingSiacoinsFloat)
		// For siafunds, need to avoid having a negative types.Currency.
		if incomingSiafunds.Cmp(outgoingSiafunds) >= 0 {
			fmt.Printf("%14v SF\n", incomingSiafunds.Sub(outgoingSiafunds))
		} else {
			fmt.Printf("-%14v SF\n", outgoingSiafunds.Sub(incomingSiafunds))
		}
	}
}

// walletunlockcmd unlocks a saved wallet
func walletunlockcmd() {
	password, err := speakeasy.Ask("Wallet password: ")
	if err != nil {
		die("Reading password failed:", err)
	}
	qs := fmt.Sprintf("encryptionpassword=%s&dictonary=%s", password, "english")
	err = post("/wallet/unlock", qs)
	if err != nil {
		die("Could not unlock wallet:", err)
	}
	fmt.Println("Wallet unlocked")
}
