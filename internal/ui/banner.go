package ui

import (
	"fmt"
)

func DisplayBanner(title string, description ...string) string {
    logo := `
             /$$                         /$$
            | $$                        | $$
  /$$$$$$$ /$$$$$$    /$$$$$$   /$$$$$$$| $$$$$$$
 /$$_____/|_  $$_/   |____  $$ /$$_____/| $$__  $$
|  $$$$$$   | $$      /$$$$$$$|  $$$$$$ | $$  \ $$
 \____  $$  | $$ /$$ /$$__  $$ \____  $$| $$  | $$
 /$$$$$$$/  |  $$$$/|  $$$$$$$ /$$$$$$$/| $$  | $$
|_______/    \___/   \_______/|_______/ |__/  |__/`

	content := fmt.Sprintf("%s\n\n%s\n", logo, title)

    if len(description) > 0 {
        content += fmt.Sprintf("%s\n", description[0])
		content += "------------------------------------------------------------------"
    }

    return content
}