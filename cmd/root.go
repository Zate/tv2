package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Root   RootConfig   `mapstructure:""`
	Status StatusConfig `mapstructure:"status"`
	Talos  TalosConfig  `mapstructure:"talos"`
	Bmc    BmcConfig    `mapstructure:"bmc"`
}

type TalosConfig struct {
	Version string `mapstructure:"version" flag:"talosVersion"`
	Name    string `mapstructure:"cluster_name" flag:"clusterName"`
	IP      string `mapstructure:"cluster_ip" flag:"clusterIP"`
	Nodes   int    `mapstructure:"cluster_nodes" flag:"clusterNodes"`
}

type BmcConfig struct {
	IP       string
	User     string
	Password string
}

type RootConfig struct {
}

type StatusConfig struct {
	Verbose bool
}

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "tv2",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
	examples and usage of using your application. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}
	logger  *slog.Logger
	cfgFile string
	config  Config
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tv2.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .tv2.yaml)")
	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	// viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	// viper.SetDefault("license", "apache")
	// logger.Info("rootCmd initialized")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config file called .tv2.yaml or .tv2.yml in the following locations
		// ./, $HOME/.tv2, $HOME
		viper.AddConfigPath(".")            // look in the current working directory first
		viper.AddConfigPath(home + "/.tv2") // look in home/.tv2 next
		viper.AddConfigPath(home)           // look in home next
		viper.SetConfigType("yaml")         // look for a yaml file
		viper.SetConfigName(".tv2")         // look for a file named .tv2.yaml
	}
	// viper.SetEnvPrefix("TV2")
	// viper.AutomaticEnv()

	// if err := viper.ReadInConfig(); err == nil {
	// 	fmt.Println("Using config file:", viper.ConfigFileUsed())
	// }

	// // Unmarshal config
	// if err := viper.Unmarshal(&config); err != nil {
	// 	fmt.Printf("Unable to decode into struct, %v", err)
	// }

	viper.SetEnvPrefix("TV2")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	// Auto-bind for all commands
	autoBindConfig(rootCmd, &config, "")
	// else if err != nil {
	// 	fmt.Println("No config file found, checking for environment variables")

	// 	if os.Getenv("TV2_CONFIG") != "" {
	// 		viper.SetConfigFile(os.Getenv("TV2_CONFIG"))
	// 		if err := viper.ReadInConfig(); err == nil {
	// 			fmt.Println("Using config file:", viper.ConfigFileUsed())
	// 		}
	// 	}
	// }

	// if err := viper.ReadInConfig(); err != nil {
	// 	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
	// 		// Config file not found; ignore error if desired
	// 	} else {
	// 		// Config file was found but another error was produced
	// 	}
	// }

	// Config file found and successfully parsed
}

func autoBindConfig(cmd *cobra.Command, config interface{}, prefix string) {
	v := reflect.ValueOf(config)
	t := v.Type()

	// If it's a pointer, get the type it points to
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	// Ensure we're working with a struct
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// If it's a nested struct, recurse
		if fieldValue.Kind() == reflect.Struct {
			autoBindConfig(cmd, fieldValue.Addr().Interface(), prefix+strings.ToLower(field.Name)+".")
			continue
		}

		flagName := field.Tag.Get("flag")
		if flagName == "" {
			continue
		}

		envName := strings.ToUpper(prefix + strings.ReplaceAll(field.Tag.Get("mapstructure"), "_", "_"))
		viperKey := prefix + field.Tag.Get("mapstructure")

		// Bind flag
		flag := cmd.Flags().Lookup(flagName)
		if flag != nil {
			err := viper.BindPFlag(viperKey, flag)
			cobra.CheckErr(err)
		}

		// Bind env
		err := viper.BindEnv(viperKey, envName)
		cobra.CheckErr(err)
	}
}
