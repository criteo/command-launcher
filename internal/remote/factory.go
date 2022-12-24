package remote

func CreateRemoteRepository(repoRootUrl string) RemoteRepository {
	return newRemoteRepository(repoRootUrl)
}
