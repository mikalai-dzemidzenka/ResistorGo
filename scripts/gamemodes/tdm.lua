-- Team deathmatch: every player is put on Red or Blue as it enters the match,
-- the smaller team first; a kill of the other side scores for the killer's
-- team, and the score is called out every minute. Bots stay on the bots' team,
-- neutral: they fight both sides, and breaking one scores nothing.
--
-- Run it with:  resistor -dedicated -gamemode tdm -bots 4

-- Teams 0 (nobody's) and 1 (the bots') are what fighters start on, so the two
-- sides are 2 and 3, and a fighter on either is one already sorted. The game
-- draws a red line under a fighter on 2 and a blue one under a fighter on 3.
local RED, BLUE = 2, 3
local names = { [RED] = "Red", [BLUE] = "Blue" }
local score = { [RED] = 0, [BLUE] = 0 }

local function teamSize(team)
  local n = 0
  for _, id in ipairs(GetFighters()) do
    if GetFighterTeam(id) == team then n = n + 1 end
  end
  return n
end

local function who(id)
  local player = GetFighterPlayer(id)
  if player and player > 0 then return "player " .. player end
  return "a bot"
end

function OnGameModeInit()
  print("team deathmatch: Red against Blue")
  SetTimer(function()
    SendMessageToAll("Score - Red " .. score[RED] .. " : " .. score[BLUE] .. " Blue")
  end, 60000, true)
end

function OnFighterSpawn(id, player)
  if player == 0 then return end -- a bot stays neutral
  -- a player coming back keeps the team it died on
  local team = GetFighterTeam(id)
  if team ~= RED and team ~= BLUE then
    team = teamSize(RED) <= teamSize(BLUE) and RED or BLUE
    SetFighterTeam(id, team)
  end
  SendPlayerMessage(player, "You fight for " .. names[team] .. ".")
end

function OnFighterDeath(id, killer)
  local lost = GetFighterTeam(id)
  local won = killer > 0 and GetFighterTeam(killer)
  if not won or won == lost or not names[won] or not names[lost] then return end
  score[won] = score[won] + 1
  SendMessageToAll(names[won] .. " scores: " .. who(killer) .. " broke " .. who(id) .. ".  Red " .. score[RED] .. " : " .. score[BLUE] .. " Blue")
end

function OnGameModeExit()
  print("final score: Red " .. score[RED] .. " : " .. score[BLUE] .. " Blue")
end
