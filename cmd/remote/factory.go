package remote

func CreateRemoteRepository(repoRootUrl string) RemoteRepository {
	return newCdtRemoteRepository(repoRootUrl)
}
