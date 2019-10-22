package discord

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"ichor-stats/src/app/models/faceit"
	"ichor-stats/src/app/services/config"
	"ichor-stats/src/app/services/discord/helpers"
	"ichor-stats/src/package/api"
	client "ichor-stats/src/package/http"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!") {
		requesterID := GetRequesterID(m.Author.ID)

		// Create a Bearer string by appending string access token
		var bearer = "Bearer " + config.GetConfig().FACEIT_API_KEY
		req, err := http.NewRequest("GET", api.GetFaceitPlayerCsgoStats(requesterID), nil)
		if err != nil {
			log.Println(err)
			return
		}

		// add authorization header to the req
		req.Header.Add("Authorization", bearer)

		response, err := client.Fire(req)
		if err != nil {
			log.Println(err)
			return
		}

		var stats faceit.Stats
		err = json.NewDecoder(response.Body).Decode(&stats)
		if err != nil {
			log.Println(err)
			return
		}

		req, err = http.NewRequest("GET", api.GetFaceitPlayerStats(requesterID), nil)
		if err != nil {
			log.Println(err)
			return
		}
		req.Header.Add("Authorization", bearer)

		response, err = client.Fire(req)
		if err != nil {
			log.Println(err)
			return
		}

		var user faceit.User
		err = json.NewDecoder(response.Body).Decode(&user)
		if err != nil {
			log.Println(err)
			return
		}

		embed := helpers.NewEmbed().
			SetTitle("Bot Invalid Command").
			AddField("UNKNOWN COMMAND",  m.Content, true)

		if strings.HasPrefix(m.Content, "!stats") {
			kills, assists, deaths := DetermineTotalStats(m.Content, stats, user)

			embed = helpers.NewEmbed().
				SetTitle(user.Games.CSGO.Name).
				AddField("ELO", strconv.Itoa(user.Games.CSGO.ELO), true).
				AddField("Skill Level", strconv.Itoa(user.Games.CSGO.SkillLevel), true).
				AddField("Avg. K/D Ratio", stats.Lifetime.AverageKD, false).
				AddField("Avg. Headshots %", stats.Lifetime.AverageHeadshots, false).
				AddField("Total Kills", kills, true).
				AddField("Total Assists", assists, true).
				AddField("Total Deaths", deaths, true)
		} else if strings.HasPrefix(m.Content, "!streak") {
			embed = helpers.NewEmbed().
				SetTitle(user.Games.CSGO.Name).
				AddField("Current Win Streak", stats.Lifetime.CurrentWinStreak, true)
		} else if strings.HasPrefix(m.Content, "!recent") {
			var resultsArray []string
			for _, result := range stats.Lifetime.RecentResults {
				if result == "0" {
					resultsArray =append(resultsArray, "L")
				} else {
					resultsArray =append(resultsArray, "W")
				}
			}

			embed = helpers.NewEmbed().
				SetTitle(user.Games.CSGO.Name).
				AddField("Recent Results (Most recent on right)", strings.Join(resultsArray, ", "), true)
		} else if strings.HasPrefix(m.Content, "!green") {
			embed = helpers.NewEmbed().
				SetTitle("World's Best Player").
				AddField("Will steal your wife and kids.",  m.Content, true)
		} else if strings.HasPrefix(m.Content, "!map") {
			embed = DetermineMapStats(m.Content, stats, user)
		}

		if embed != nil {
			_, err = s.ChannelMessageSendEmbed(config.GetConfig().CHANNEL_ID, embed.MessageEmbed)
		}

		if err != nil {
			log.Println(err)
		}
	}
}

func GetRequesterID(discordID string) string {
	if discordID == "210457267710066689" {
		return "0d94613d-b736-46ba-b8cd-d2159ddad705"
	} else if discordID == "210449893892947969" {
		return "b26df7d4-8517-4ec6-ab58-708487e5fe60"
	} else if discordID == "210438278623526913" {
		return "b0a57a5a-2f7a-481c-aaa8-8013a83378e3"
	}

	return ""
}

func DetermineMapStats(data string, stats faceit.Stats, user faceit.User) *helpers.Embed {
	mapString := strings.Split(data, " ")

	if len(mapString) > 1 {
		for _, result := range stats.Segment {
			kills, _ := strconv.ParseFloat(result.LifetimeMapStats.Kills, 64)
			deaths, _ := strconv.ParseFloat(result.LifetimeMapStats.Deaths, 64)

			if strings.HasSuffix(result.CsMap, mapString[1]) {
				return helpers.NewEmbed().
					SetTitle("Map statistics for " + user.Games.CSGO.Name).
					AddField("Map",  result.CsMap, false).
					AddField("Kills",  result.LifetimeMapStats.Kills, true).
					AddField("Assists",  result.LifetimeMapStats.Assists, true).
					AddField("Deaths",  result.LifetimeMapStats.Deaths, true).
					AddField("K/D Ratio", strconv.FormatFloat(kills/deaths, 'f', 2, 64), false)
			}
		}
	}

	return nil
}

func DetermineTotalStats(data string, stats faceit.Stats, user faceit.User) (string, string, string) {
	totalKills := 0
	totalDeaths := 0
	totalAssists := 0

	for _, result := range stats.Segment {
		mapKills, _ := strconv.Atoi(result.LifetimeMapStats.Kills)
		mapDeaths, _ := strconv.Atoi(result.LifetimeMapStats.Deaths)
		mapAssists, _ := strconv.Atoi(result.LifetimeMapStats.Assists)

		totalKills = totalKills + mapKills
		totalDeaths = totalDeaths + mapDeaths
		totalAssists = totalAssists + mapAssists
	}

	return strconv.Itoa(totalKills), strconv.Itoa(totalAssists), strconv.Itoa(totalDeaths)
}