-- Greets every player who joins, and tells everyone when one leaves.
--
-- Run it alongside any gamemode:  resistor -dedicated -filterscripts welcome

function OnFilterScriptInit()
  print("welcome loaded")
end

function OnPlayerConnect(player)
  SendPlayerMessage(player, "Welcome to Resistor, player " .. player .. ".")
  SendMessageToAll("Player " .. player .. " has joined.")
end

function OnPlayerDisconnect(player, reason)
  local why = { quit = "left", timeout = "timed out", kicked = "was kicked" }
  SendMessageToAll("Player " .. player .. " " .. (why[reason] or "left") .. ".")
end
