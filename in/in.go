// Source generated for release WIN63-202312011715-920410468 (source: Sulek API)

package in

import g "github.com/b7c/goearth"

func id(name string) g.Identifier {
	return g.Identifier{Dir: g.In, Name: name}
}

var (
	PromoArticles = id("PromoArticles")
	FurniListAddOrUpdate = id("FurniListAddOrUpdate")
	FurniList = id("FurniList")
	FurniListInvalidate = id("FurniListInvalidate")
	FurniListRemove = id("FurniListRemove")
	PostItPlaced = id("PostItPlaced")
	CanCreateRoom = id("CanCreateRoom")
	CanCreateRoomEvent = id("CanCreateRoomEvent")
	CategoriesWithVisitorCount = id("CategoriesWithVisitorCount")
	CompetitionRoomsData = id("CompetitionRoomsData")
	ConvertedRoomId = id("ConvertedRoomId")
	Doorbell = id("Doorbell")
	FavouriteChanged = id("FavouriteChanged")
	Favourites = id("Favourites")
	FlatAccessDenied = id("FlatAccessDenied")
	FlatCreated = id("FlatCreated")
	GetGuestRoomResult = id("GetGuestRoomResult")
	GuestRoomSearchResult = id("GuestRoomSearchResult")
	NavigatorSettings = id("NavigatorSettings")
	OfficialRooms = id("OfficialRooms")
	PopularRoomTagsResult = id("PopularRoomTagsResult")
	RoomEventCancel = id("RoomEventCancel")
	RoomEvent = id("RoomEvent")
	RoomInfoUpdated = id("RoomInfoUpdated")
	RoomRating = id("RoomRating")
	UserEventCats = id("UserEventCats")
	UserFlatCats = id("UserFlatCats")
	CommunityVoteReceived = id("CommunityVoteReceived")
	ChangeUserNameResult = id("ChangeUserNameResult")
	CheckUserNameResult = id("CheckUserNameResult")
	FigureUpdate = id("FigureUpdate")
	Wardrobe = id("Wardrobe")
	RoomEntryTile = id("RoomEntryTile")
	RoomOccupiedTiles = id("RoomOccupiedTiles")
	BotCommandConfiguration = id("BotCommandConfiguration")
	BotError = id("BotError")
	BotForceOpenContextMenu = id("BotForceOpenContextMenu")
	BotSkillListUpdate = id("BotSkillListUpdate")
	AvatarEffect = id("AvatarEffect")
	CarryObject = id("CarryObject")
	Dance = id("Dance")
	Expression = id("Expression")
	Sleep = id("Sleep")
	UseObject = id("UseObject")
	TalentLevelUp = id("TalentLevelUp")
	TalentTrackLevel = id("TalentTrackLevel")
	TalentTrack = id("TalentTrack")
	Game2FriendsLeaderboard = id("Game2FriendsLeaderboard")
	Game2TotalGroupLeaderboard = id("Game2TotalGroupLeaderboard")
	Game2TotalLeaderboard = id("Game2TotalLeaderboard")
	Game2WeeklyFriendsLeaderboard = id("Game2WeeklyFriendsLeaderboard")
	Game2WeeklyGroupLeaderboard = id("Game2WeeklyGroupLeaderboard")
	Game2WeeklyLeaderboard = id("Game2WeeklyLeaderboard")
	Chat = id("Chat")
	FloodControl = id("FloodControl")
	RemainingMutePeriod = id("RemainingMutePeriod")
	RoomChatSettings = id("RoomChatSettings")
	RoomFilterSettings = id("RoomFilterSettings")
	Shout = id("Shout")
	UserTyping = id("UserTyping")
	Whisper = id("Whisper")
	CfhChatlog = id("CfhChatlog")
	IssueDeleted = id("IssueDeleted")
	IssueInfo = id("IssueInfo")
	IssuePickFailed = id("IssuePickFailed")
	ModeratorActionResult = id("ModeratorActionResult")
	ModeratorCaution = id("ModeratorCaution")
	ModeratorInit = id("ModeratorInit")
	Moderator = id("Moderator")
	ModeratorRoomInfo = id("ModeratorRoomInfo")
	ModeratorToolPreferences = id("ModeratorToolPreferences")
	ModeratorUserInfo = id("ModeratorUserInfo")
	RoomChatlog = id("RoomChatlog")
	RoomVisits = id("RoomVisits")
	UserBanned = id("UserBanned")
	UserChatlog = id("UserChatlog")
	CfhSanction = id("CfhSanction")
	CfhTopicsInit = id("CfhTopicsInit")
	SanctionStatus = id("SanctionStatus")
	Game2FullGameStatus = id("Game2FullGameStatus")
	Game2GameStatus = id("Game2GameStatus")
	Game2AccountGameStatus = id("Game2AccountGameStatus")
	Game2GameCancelled = id("Game2GameCancelled")
	Game2GameCreated = id("Game2GameCreated")
	Game2GameDirectoryStatus = id("Game2GameDirectoryStatus")
	Game2GameLongData = id("Game2GameLongData")
	Game2GameStarted = id("Game2GameStarted")
	Game2InArenaQueue = id("Game2InArenaQueue")
	Game2JoiningGameFailed = id("Game2JoiningGameFailed")
	Game2StartCounter = id("Game2StartCounter")
	Game2StartingGameFailed = id("Game2StartingGameFailed")
	Game2StopCounter = id("Game2StopCounter")
	Game2UserBlocked = id("Game2UserBlocked")
	Game2UserJoinedGame = id("Game2UserJoinedGame")
	Game2UserLeftGame = id("Game2UserLeftGame")
	AvatarEffectActivated = id("AvatarEffectActivated")
	AvatarEffectAdded = id("AvatarEffectAdded")
	AvatarEffectExpired = id("AvatarEffectExpired")
	AvatarEffectSelected = id("AvatarEffectSelected")
	AvatarEffects = id("AvatarEffects")
	AcceptFriendResult = id("AcceptFriendResult")
	ConsoleMessageHistory = id("ConsoleMessageHistory")
	FindFriendsProcessResult = id("FindFriendsProcessResult")
	FollowFriendFailed = id("FollowFriendFailed")
	FriendListFragment = id("FriendListFragment")
	FriendListUpdate = id("FriendListUpdate")
	FriendNotification = id("FriendNotification")
	FriendRequests = id("FriendRequests")
	HabboSearchResult = id("HabboSearchResult")
	InstantMessageError = id("InstantMessageError")
	MessengerError = id("MessengerError")
	MessengerInit = id("MessengerInit")
	MiniMailNew = id("MiniMailNew")
	MiniMailUnreadCount = id("MiniMailUnreadCount")
	NewConsole = id("NewConsole")
	NewFriendRequest = id("NewFriendRequest")
	RoomInviteError = id("RoomInviteError")
	RoomInvite = id("RoomInvite")
	PetBreedingResult = id("PetBreedingResult")
	PetCommands = id("PetCommands")
	PetExperience = id("PetExperience")
	PetFigureUpdate = id("PetFigureUpdate")
	PetInfo = id("PetInfo")
	PetLevelUpdate = id("PetLevelUpdate")
	PetPlacingError = id("PetPlacingError")
	PetRespectFailed = id("PetRespectFailed")
	PetStatusUpdate = id("PetStatusUpdate")
	AvailabilityStatus = id("AvailabilityStatus")
	InfoHotelClosed = id("InfoHotelClosed")
	InfoHotelClosing = id("InfoHotelClosing")
	LoginFailedHotelClosed = id("LoginFailedHotelClosed")
	MaintenanceStatus = id("MaintenanceStatus")
	ConfirmBreedingRequest = id("ConfirmBreedingRequest")
	ConfirmBreedingResult = id("ConfirmBreedingResult")
	GoToBreedingNestFailure = id("GoToBreedingNestFailure")
	NestBreedingSuccess = id("NestBreedingSuccess")
	PetAddedToInventory = id("PetAddedToInventory")
	PetBreeding = id("PetBreeding")
	PetInventory = id("PetInventory")
	PetReceived = id("PetReceived")
	PetRemovedFromInventory = id("PetRemovedFromInventory")
	UserNftWardrobe = id("UserNftWardrobe")
	UserNftWardrobeSelection = id("UserNftWardrobeSelection")
	CantConnect = id("CantConnect")
	CloseConnection = id("CloseConnection")
	FlatAccessible = id("FlatAccessible")
	GamePlayerValue = id("GamePlayerValue")
	HanditemConfiguration = id("HanditemConfiguration")
	OpenConnection = id("OpenConnection")
	RoomForward = id("RoomForward")
	RoomQueueStatus = id("RoomQueueStatus")
	RoomReady = id("RoomReady")
	YouArePlayingGame = id("YouArePlayingGame")
	YouAreSpectator = id("YouAreSpectator")
	BadgePointLimits = id("BadgePointLimits")
	BadgeReceived = id("BadgeReceived")
	Badges = id("Badges")
	IsBadgeRequestFulfilled = id("IsBadgeRequestFulfilled")
	CampaignCalendarData = id("CampaignCalendarData")
	CampaignCalendarDoorOpened = id("CampaignCalendarDoorOpened")
	AccountPreferences = id("AccountPreferences")
	CreditVaultStatus = id("CreditVaultStatus")
	IncomeRewardClaimResponse = id("IncomeRewardClaimResponse")
	IncomeRewardStatus = id("IncomeRewardStatus")
	BotAddedToInventory = id("BotAddedToInventory")
	BotInventory = id("BotInventory")
	BotRemovedFromInventory = id("BotRemovedFromInventory")
	CancelMysteryBoxWait = id("CancelMysteryBoxWait")
	GotMysteryBoxPrize = id("GotMysteryBoxPrize")
	MysteryBoxKeys = id("MysteryBoxKeys")
	ShowMysteryBoxWait = id("ShowMysteryBoxWait")
	PollContents = id("PollContents")
	PollError = id("PollError")
	PollOffer = id("PollOffer")
	QuestionAnswered = id("QuestionAnswered")
	Question = id("Question")
	QuestionFinished = id("QuestionFinished")
	TradeOpenFailed = id("TradeOpenFailed")
	TradingAccept = id("TradingAccept")
	TradingClose = id("TradingClose")
	TradingCompleted = id("TradingCompleted")
	TradingConfirmation = id("TradingConfirmation")
	TradingItemList = id("TradingItemList")
	TradingNotOpen = id("TradingNotOpen")
	TradingOpen = id("TradingOpen")
	TradingOtherNotAllowed = id("TradingOtherNotAllowed")
	TradingYouAreNotAllowed = id("TradingYouAreNotAllowed")
	MarketplaceBuyOfferResult = id("MarketplaceBuyOfferResult")
	MarketplaceCancelOfferResult = id("MarketplaceCancelOfferResult")
	MarketplaceCanMakeOfferResult = id("MarketplaceCanMakeOfferResult")
	MarketplaceConfiguration = id("MarketplaceConfiguration")
	MarketplaceItemStats = id("MarketplaceItemStats")
	MarketplaceMakeOfferResult = id("MarketplaceMakeOfferResult")
	MarketPlaceOffers = id("MarketPlaceOffers")
	MarketPlaceOwnOffers = id("MarketPlaceOwnOffers")
	ErrorReport = id("ErrorReport")
	FigureSetIds = id("FigureSetIds")
	AccountSafetyLockStatusChange = id("AccountSafetyLockStatusChange")
	ApproveName = id("ApproveName")
	ChangeEmailResult = id("ChangeEmailResult")
	EmailStatusResult = id("EmailStatusResult")
	ExtendedProfileChanged = id("ExtendedProfileChanged")
	ExtendedProfile = id("ExtendedProfile")
	GroupDetailsChanged = id("GroupDetailsChanged")
	GroupMembershipRequested = id("GroupMembershipRequested")
	GuildCreated = id("GuildCreated")
	GuildCreationInfo = id("GuildCreationInfo")
	GuildEditFailed = id("GuildEditFailed")
	GuildEditInfo = id("GuildEditInfo")
	GuildEditorData = id("GuildEditorData")
	GuildMemberFurniCountInHQ = id("GuildMemberFurniCountInHQ")
	GuildMemberMgmtFailed = id("GuildMemberMgmtFailed")
	GuildMembershipRejected = id("GuildMembershipRejected")
	GuildMemberships = id("GuildMemberships")
	GuildMembershipUpdated = id("GuildMembershipUpdated")
	GuildMembers = id("GuildMembers")
	HabboGroupBadges = id("HabboGroupBadges")
	HabboGroupDeactivated = id("HabboGroupDeactivated")
	HabboGroupDetails = id("HabboGroupDetails")
	HabboGroupJoinFailed = id("HabboGroupJoinFailed")
	HabboUserBadges = id("HabboUserBadges")
	HandItemReceived = id("HandItemReceived")
	IgnoredUsers = id("IgnoredUsers")
	IgnoreResult = id("IgnoreResult")
	InClientLink = id("InClientLink")
	PetRespectNotification = id("PetRespectNotification")
	PetSupplementedNotification = id("PetSupplementedNotification")
	RelationshipStatusInfo = id("RelationshipStatusInfo")
	RespectNotification = id("RespectNotification")
	ScrSendKickbackInfo = id("ScrSendKickbackInfo")
	ScrSendUserInfo = id("ScrSendUserInfo")
	UserNameChanged = id("UserNameChanged")
	AchievementResolutionCompleted = id("AchievementResolutionCompleted")
	AchievementResolutionProgress = id("AchievementResolutionProgress")
	AchievementResolutions = id("AchievementResolutions")
	CreditBalance = id("CreditBalance")
	BonusRareInfo = id("BonusRareInfo")
	BuildersClubFurniCount = id("BuildersClubFurniCount")
	BuildersClubSubscriptionStatus = id("BuildersClubSubscriptionStatus")
	BundleDiscountRuleset = id("BundleDiscountRuleset")
	CatalogIndex = id("CatalogIndex")
	CatalogPage = id("CatalogPage")
	CatalogPageWithEarliestExpiry = id("CatalogPageWithEarliestExpiry")
	CatalogPublished = id("CatalogPublished")
	ClubGiftInfo = id("ClubGiftInfo")
	ClubGiftSelected = id("ClubGiftSelected")
	GiftReceiverNotFound = id("GiftReceiverNotFound")
	GiftWrappingConfiguration = id("GiftWrappingConfiguration")
	HabboClubExtendOffer = id("HabboClubExtendOffer")
	HabboClubOffers = id("HabboClubOffers")
	LimitedEditionSoldOut = id("LimitedEditionSoldOut")
	LimitedOfferAppearingNext = id("LimitedOfferAppearingNext")
	NotEnoughBalance = id("NotEnoughBalance")
	ProductOffer = id("ProductOffer")
	PurchaseError = id("PurchaseError")
	PurchaseNotAllowed = id("PurchaseNotAllowed")
	PurchaseOK = id("PurchaseOK")
	RoomAdPurchaseInfo = id("RoomAdPurchaseInfo")
	SeasonalCalendarDailyOffer = id("SeasonalCalendarDailyOffer")
	SellablePetPalettes = id("SellablePetPalettes")
	SnowWarGameTokens = id("SnowWarGameTokens")
	TargetedOffer = id("TargetedOffer")
	TargetedOfferNotFound = id("TargetedOfferNotFound")
	VoucherRedeemError = id("VoucherRedeemError")
	VoucherRedeemOk = id("VoucherRedeemOk")
	CameraPublishStatus = id("CameraPublishStatus")
	CameraPurchaseOK = id("CameraPurchaseOK")
	CameraStorageUrl = id("CameraStorageUrl")
	CompetitionStatus = id("CompetitionStatus")
	InitCamera = id("InitCamera")
	ThumbnailStatus = id("ThumbnailStatus")
	PhoneCollectionState = id("PhoneCollectionState")
	TryPhoneNumberResult = id("TryPhoneNumberResult")
	TryVerificationCodeResult = id("TryVerificationCodeResult")
	CustomStackingHeightUpdate = id("CustomStackingHeightUpdate")
	CustomUserNotification = id("CustomUserNotification")
	DiceValue = id("DiceValue")
	FurniRentOrBuyoutOffer = id("FurniRentOrBuyoutOffer")
	GuildFurniContextMenuInfo = id("GuildFurniContextMenuInfo")
	OneWayDoorStatus = id("OneWayDoorStatus")
	OpenPetPackageRequested = id("OpenPetPackageRequested")
	OpenPetPackageResult = id("OpenPetPackageResult")
	PresentOpened = id("PresentOpened")
	RentableSpaceRentFailed = id("RentableSpaceRentFailed")
	RentableSpaceRentOk = id("RentableSpaceRentOk")
	RentableSpaceStatus = id("RentableSpaceStatus")
	RequestSpamWallPostIt = id("RequestSpamWallPostIt")
	RoomDimmerPresets = id("RoomDimmerPresets")
	RoomMessageNotification = id("RoomMessageNotification")
	YoutubeControlVideo = id("YoutubeControlVideo")
	YoutubeDisplayPlaylists = id("YoutubeDisplayPlaylists")
	YoutubeDisplayVideo = id("YoutubeDisplayVideo")
	Interstitial = id("Interstitial")
	RoomAdError = id("RoomAdError")
	Achievement = id("Achievement")
	Achievements = id("Achievements")
	AchievementsScore = id("AchievementsScore")
	BannedUsersFromRoom = id("BannedUsersFromRoom")
	FlatControllerAdded = id("FlatControllerAdded")
	FlatControllerRemoved = id("FlatControllerRemoved")
	FlatControllers = id("FlatControllers")
	MuteAllInRoom = id("MuteAllInRoom")
	NoSuchFlat = id("NoSuchFlat")
	RoomSettingsData = id("RoomSettingsData")
	RoomSettingsError = id("RoomSettingsError")
	RoomSettingsSaved = id("RoomSettingsSaved")
	RoomSettingsSaveError = id("RoomSettingsSaveError")
	ShowEnforceRoomCategoryDialog = id("ShowEnforceRoomCategoryDialog")
	UserUnbannedFromRoom = id("UserUnbannedFromRoom")
	HotLooks = id("HotLooks")
	BuildersClubPlacementWarning = id("BuildersClubPlacementWarning")
	FavoriteMembershipUpdate = id("FavoriteMembershipUpdate")
	FloorHeightMap = id("FloorHeightMap")
	FurnitureAliases = id("FurnitureAliases")
	HeightMap = id("HeightMap")
	HeightMapUpdate = id("HeightMapUpdate")
	ItemAdd = id("ItemAdd")
	ItemDataUpdate = id("ItemDataUpdate")
	ItemRemove = id("ItemRemove")
	Items = id("Items")
	ItemsStateUpdate = id("ItemsStateUpdate")
	ItemStateUpdate = id("ItemStateUpdate")
	ItemUpdate = id("ItemUpdate")
	ObjectAdd = id("ObjectAdd")
	ObjectDataUpdate = id("ObjectDataUpdate")
	ObjectRemove = id("ObjectRemove")
	ObjectRemoveMultiple = id("ObjectRemoveMultiple")
	ObjectsDataUpdate = id("ObjectsDataUpdate")
	Objects = id("Objects")
	ObjectUpdate = id("ObjectUpdate")
	RoomEntryInfo = id("RoomEntryInfo")
	RoomProperty = id("RoomProperty")
	RoomVisualizationSettings = id("RoomVisualizationSettings")
	SlideObjectBundle = id("SlideObjectBundle")
	SpecialRoomEffect = id("SpecialRoomEffect")
	UserChange = id("UserChange")
	UserRemove = id("UserRemove")
	Users = id("Users")
	UserUpdate = id("UserUpdate")
	WiredMovements = id("WiredMovements")
	FriendFurniCancelLock = id("FriendFurniCancelLock")
	FriendFurniOtherLockConfirmed = id("FriendFurniOtherLockConfirmed")
	FriendFurniStartConfirmation = id("FriendFurniStartConfirmation")
	NavigatorCollapsedCategories = id("NavigatorCollapsedCategories")
	NavigatorLiftedRooms = id("NavigatorLiftedRooms")
	NavigatorMetaData = id("NavigatorMetaData")
	NavigatorSavedSearches = id("NavigatorSavedSearches")
	NavigatorSearchResultBlocks = id("NavigatorSearchResultBlocks")
	NewNavigatorPreferences = id("NewNavigatorPreferences")
	UserClassification = id("UserClassification")
	ForumData = id("ForumData")
	ForumsList = id("ForumsList")
	ForumThreads = id("ForumThreads")
	PostMessage = id("PostMessage")
	PostThread = id("PostThread")
	ThreadMessages = id("ThreadMessages")
	UnreadForumsCount = id("UnreadForumsCount")
	UpdateMessage = id("UpdateMessage")
	UpdateThread = id("UpdateThread")
	CallForHelpDisabledNotify = id("CallForHelpDisabledNotify")
	CallForHelpPendingCallsDeleted = id("CallForHelpPendingCallsDeleted")
	CallForHelpPendingCalls = id("CallForHelpPendingCalls")
	CallForHelpReply = id("CallForHelpReply")
	CallForHelpResult = id("CallForHelpResult")
	ChatReviewSessionDetached = id("ChatReviewSessionDetached")
	ChatReviewSessionOfferedToGuide = id("ChatReviewSessionOfferedToGuide")
	ChatReviewSessionResults = id("ChatReviewSessionResults")
	ChatReviewSessionStarted = id("ChatReviewSessionStarted")
	ChatReviewSessionVotingStatus = id("ChatReviewSessionVotingStatus")
	GuideOnDutyStatus = id("GuideOnDutyStatus")
	GuideReportingStatus = id("GuideReportingStatus")
	GuideSessionAttached = id("GuideSessionAttached")
	GuideSessionDetached = id("GuideSessionDetached")
	GuideSessionEnded = id("GuideSessionEnded")
	GuideSessionError = id("GuideSessionError")
	GuideSessionInvitedToGuideRoom = id("GuideSessionInvitedToGuideRoom")
	GuideSessionMessage = id("GuideSessionMessage")
	GuideSessionPartnerIsTyping = id("GuideSessionPartnerIsTyping")
	GuideSessionRequesterRoom = id("GuideSessionRequesterRoom")
	GuideSessionStarted = id("GuideSessionStarted")
	GuideTicketCreationResult = id("GuideTicketCreationResult")
	GuideTicketResolution = id("GuideTicketResolution")
	IssueCloseNotification = id("IssueCloseNotification")
	QuizData = id("QuizData")
	QuizResults = id("QuizResults")
	ActivityPoints = id("ActivityPoints")
	ClubGiftNotification = id("ClubGiftNotification")
	ElementPointer = id("ElementPointer")
	HabboAchievementNotification = id("HabboAchievementNotification")
	HabboActivityPointNotification = id("HabboActivityPointNotification")
	HabboBroadcast = id("HabboBroadcast")
	InfoFeedEnable = id("InfoFeedEnable")
	MOTDNotification = id("MOTDNotification")
	NotificationDialog = id("NotificationDialog")
	OfferRewardDelivered = id("OfferRewardDelivered")
	PetLevelNotification = id("PetLevelNotification")
	RestoreClient = id("RestoreClient")
	UnseenItems = id("UnseenItems")
	Open = id("Open")
	WiredFurniAction = id("WiredFurniAction")
	WiredFurniAddon = id("WiredFurniAddon")
	WiredFurniCondition = id("WiredFurniCondition")
	WiredFurniSelector = id("WiredFurniSelector")
	WiredFurniTrigger = id("WiredFurniTrigger")
	WiredRewardResult = id("WiredRewardResult")
	WiredSaveSuccess = id("WiredSaveSuccess")
	WiredValidationError = id("WiredValidationError")
	Game2ArenaEntered = id("Game2ArenaEntered")
	Game2EnterArenaFailed = id("Game2EnterArenaFailed")
	Game2EnterArena = id("Game2EnterArena")
	Game2GameChatFromPlayer = id("Game2GameChatFromPlayer")
	Game2GameEnding = id("Game2GameEnding")
	Game2GameRejoin = id("Game2GameRejoin")
	Game2PlayerExitedGameArena = id("Game2PlayerExitedGameArena")
	Game2PlayerRematches = id("Game2PlayerRematches")
	Game2StageEnding = id("Game2StageEnding")
	Game2StageLoad = id("Game2StageLoad")
	Game2StageRunning = id("Game2StageRunning")
	Game2StageStarting = id("Game2StageStarting")
	Game2StageStillLoading = id("Game2StageStillLoading")
	YouAreController = id("YouAreController")
	YouAreNotController = id("YouAreNotController")
	YouAreOwner = id("YouAreOwner")
	JukeboxPlayListFull = id("JukeboxPlayListFull")
	JukeboxSongDisks = id("JukeboxSongDisks")
	NowPlaying = id("NowPlaying")
	OfficialSongId = id("OfficialSongId")
	PlayList = id("PlayList")
	PlayListSongAdded = id("PlayListSongAdded")
	TraxSongInfo = id("TraxSongInfo")
	UserSongDisksInventory = id("UserSongDisksInventory")
	NewUserExperienceGiftOffer = id("NewUserExperienceGiftOffer")
	NewUserExperienceNotComplete = id("NewUserExperienceNotComplete")
	SelectInitialRoom = id("SelectInitialRoom")
	CommunityGoalHallOfFame = id("CommunityGoalHallOfFame")
	CommunityGoalProgress = id("CommunityGoalProgress")
	ConcurrentUsersGoalProgress = id("ConcurrentUsersGoalProgress")
	EpicPopup = id("EpicPopup")
	QuestCancelled = id("QuestCancelled")
	QuestCompleted = id("QuestCompleted")
	QuestDaily = id("QuestDaily")
	Quest = id("Quest")
	Quests = id("Quests")
	SeasonalQuests = id("SeasonalQuests")
	CitizenshipVipOfferPromoEnabled = id("CitizenshipVipOfferPromoEnabled")
	PerkAllowances = id("PerkAllowances")
	CompetitionEntrySubmitResult = id("CompetitionEntrySubmitResult")
	CompetitionVotingInfo = id("CompetitionVotingInfo")
	CurrentTimingCode = id("CurrentTimingCode")
	IsUserPartOfCompetition = id("IsUserPartOfCompetition")
	NoOwnedRoomsAlert = id("NoOwnedRoomsAlert")
	SecondsUntil = id("SecondsUntil")
	LatencyPingResponse = id("LatencyPingResponse")
	CraftableProducts = id("CraftableProducts")
	CraftingRecipe = id("CraftingRecipe")
	CraftingRecipesAvailable = id("CraftingRecipesAvailable")
	CraftingResult = id("CraftingResult")
	AuthenticationOK = id("AuthenticationOK")
	CompleteDiffieHandshake = id("CompleteDiffieHandshake")
	DisconnectReason = id("DisconnectReason")
	GenericError = id("GenericError")
	IdentityAccounts = id("IdentityAccounts")
	InitDiffieHandshake = id("InitDiffieHandshake")
	IsFirstLoginOfDay = id("IsFirstLoginOfDay")
	NoobnessLevel = id("NoobnessLevel")
	Ping = id("Ping")
	UniqueMachineID = id("UniqueMachineID")
	UserObject = id("UserObject")
	UserRights = id("UserRights")
)
