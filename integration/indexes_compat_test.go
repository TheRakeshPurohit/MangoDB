// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

func TestListIndexesCompat(t *testing.T) {
	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                shareddata.AllProviders(),
		AddNonExistentCollection: true,
	})
	ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

	for i := range targetCollections {
		targetCollection := targetCollections[i]
		compatCollection := compatCollections[i]

		t.Run(targetCollection.Name(), func(t *testing.T) {
			t.Helper()
			t.Parallel()

			targetCursor, targetErr := targetCollection.Indexes().List(ctx)
			compatCursor, compatErr := compatCollection.Indexes().List(ctx)

			if targetCursor != nil {
				defer targetCursor.Close(ctx)
			}
			if compatCursor != nil {
				defer compatCursor.Close(ctx)
			}

			require.NoError(t, targetErr)
			require.NoError(t, compatErr)

			targetRes := FetchAll(t, ctx, targetCursor)
			compatRes := FetchAll(t, ctx, compatCursor)

			assert.Equal(t, compatRes, targetRes)

			// Also test specifications to check they are identical.
			targetSpec, targetErr := targetCollection.Indexes().ListSpecifications(ctx)
			compatSpec, compatErr := compatCollection.Indexes().ListSpecifications(ctx)

			require.NoError(t, compatErr)
			require.NoError(t, targetErr)

			assert.Equal(t, compatSpec, targetSpec)
		})
	}
}

func TestCreateIndexesCompat(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // for readability
		models     []mongo.IndexModel
		resultType CompatTestCaseResultType // defaults to NonEmptyResult

		skip             string // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1086
		failsForFerretDB string
	}{
		"Empty": {
			models:     []mongo.IndexModel{},
			resultType: EmptyResult,
		},
		"SingleIndex": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
			},
		},
		"SingleIndexMultiField": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"foo", 1}, {"bar", -1}}},
			},
		},
		"DescendingID": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"_id", -1}}},
			},
			resultType:       EmptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/297",
		},
		"NonExistentField": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"field-does-not-exist", 1}}},
			},
		},
		"DotNotation": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v.foo", 1}}},
			},
		},
		"DangerousKey": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{
						{"v", 1},
						{"foo'))); DROP TABlE test._ferretdb_database_metadata; CREATE INDEX IF NOT EXISTS test ON test.test (((_jsonb->'foo", 1},
					},
				},
			},
		},
		"SameKey": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}, {"v", 1}}},
			},
			resultType:       EmptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/297",
		},
		"CustomName": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"foo", 1}, {"bar", -1}},
					Options: options.Index().SetName("custom-name"),
				},
			},
		},

		"MultiDirectionDifferentIndexes": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
				{Keys: bson.D{{"v", 1}}},
			},
		},
		"MultiOrder": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"foo", -1}}},
				{Keys: bson.D{{"v", 1}}},
				{Keys: bson.D{{"bar", 1}}},
			},
		},
		"MultiSameKeyUsed": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"foo", 1}}},
				{Keys: bson.D{{"foo", 1}, {"v", 1}}},
				{Keys: bson.D{{"bar", 1}}},
			},
		},
		"BuildSameIndex": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", 1}}},
				{Keys: bson.D{{"v", 1}}},
			},
			resultType: EmptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/2910",
			// the error for existing and non-existing collection are different,
			// below is the error for existing collection.
			//
			// &mongo.CommandError{
			//	Code: 96,
			//	Name: "OperationFailed",
			//	Message: `Index build failed: 7a1c4cc3-8ac6-44d3-92e0-57853e6bc837: Collection ` +
			//		`TestCreateIndexesCommandInvalidSpec-SameIndex.TestCreateIndexesCommandInvalidSpec-SameIndex ` +
			//		`( 020f17e0-7847-45f2-8397-c631c5e9bdaf ) :: caused by :: Cannot build two identical indexes. ` +
			//		`Try again without duplicate indexes.`,
			// },
		},
		"MultiWithInvalid": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{{"foo", 1}, {"bar", 1}, {"v", -1}},
				},
				{
					Keys: bson.D{{"v", -1}, {"v", 1}},
				},
			},
			resultType:       EmptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/297",
		},
		"SameKeyDifferentNames": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"v", -1}},
					Options: options.Index().SetName("foo"),
				},
				{
					Keys:    bson.D{{"v", -1}},
					Options: options.Index().SetName("bar"),
				},
			},
			resultType: EmptyResult,
		},
		"SameNameDifferentKeys": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"foo", -1}},
					Options: options.Index().SetName("index-name"),
				},
				{
					Keys:    bson.D{{"bar", -1}},
					Options: options.Index().SetName("index-name"),
				},
			},
			resultType: EmptyResult,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			// Use per-test setup because createIndexes modifies collection state,
			// however, we don't need to run index creation test for all the possible collections.
			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
				Providers:                []shareddata.Provider{shareddata.Composites},
				AddNonExistentCollection: true,
			})
			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				t.Run(targetCollection.Name(), func(tt *testing.T) {
					var t testing.TB = tt
					if tc.failsForFerretDB != "" {
						t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
					}

					targetRes, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.models)
					compatRes, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.models)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						t.Logf("Compat error: %v", compatErr)

						// error messages are intentionally not compared
						AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					assert.Equal(t, compatRes, targetRes)

					if compatErr == nil {
						nonEmptyResults = true
					}

					// List indexes to check they are identical after creation.
					targetCursor, targetErr := targetCollection.Indexes().List(ctx)
					compatCursor, compatErr := compatCollection.Indexes().List(ctx)

					if targetCursor != nil {
						defer targetCursor.Close(ctx)
					}
					if compatCursor != nil {
						defer compatCursor.Close(ctx)
					}

					require.NoError(t, targetErr)
					require.NoError(t, compatErr)

					targetIndexes := FetchAll(t, ctx, targetCursor)
					compatIndexes := FetchAll(t, ctx, compatCursor)

					assert.ElementsMatch(t, compatIndexes, targetIndexes)

					// List specifications to check they are identical after creation.
					targetSpec, targetErr := targetCollection.Indexes().ListSpecifications(ctx)
					compatSpec, compatErr := compatCollection.Indexes().ListSpecifications(ctx)

					require.NoError(t, compatErr)
					require.NoError(t, targetErr)

					require.NotEmpty(t, compatSpec)
					assert.ElementsMatch(t, compatSpec, targetSpec)
				})
			}

			switch tc.resultType {
			case NonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case EmptyResult:
				assert.False(t, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

func TestDropIndexesCompat(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // for readability
		dropIndexName    string                   // name of a single index to drop
		dropAll          bool                     // set true for drop all indexes, if true dropIndexName must be empty.
		resultType       CompatTestCaseResultType // defaults to NonEmptyResult
		toCreate         []mongo.IndexModel       // optional, if not nil create indexes before dropping
		failsForFerretDB string
	}{
		"DropAllCommand": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", 1}}},
				{Keys: bson.D{{"foo", -1}}},
				{Keys: bson.D{{"bar", 1}}},
				{Keys: bson.D{{"pam.pam", -1}}},
			},
			dropAll: true,
		},
		"ID": {
			dropIndexName: "_id_",
			resultType:    EmptyResult,
		},
		"AscendingValue": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", 1}}},
			},
			dropIndexName: "v_1",
		},
		"DescendingValue": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
			},
			dropIndexName: "v_-1",
		},
		"NonExistent": {
			dropIndexName: "nonexistent_1",
			resultType:    EmptyResult,
		},
		"Empty": {
			dropIndexName: "",
			resultType:    EmptyResult,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.dropAll {
				require.Empty(t, tc.dropIndexName, "index name must be empty when dropping all indexes")
			}

			// It's enough to use a single provider for drop indexes test as indexes work the same for different collections.
			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
				Providers:                []shareddata.Provider{shareddata.Composites},
				AddNonExistentCollection: true,
			})
			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				t.Run(targetCollection.Name(), func(tt *testing.T) {
					var t testing.TB = tt

					if tc.failsForFerretDB != "" {
						t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
					}

					if tc.toCreate != nil {
						_, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.toCreate)
						_, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.toCreate)
						require.NoError(t, compatErr)
						require.NoError(t, targetErr)
					}

					var targetRes, compatRes bson.Raw
					var targetErr, compatErr error

					if tc.dropAll {
						targetRes, targetErr = targetCollection.Indexes().DropAll(ctx)
						compatRes, compatErr = compatCollection.Indexes().DropAll(ctx)
					} else {
						targetRes, targetErr = targetCollection.Indexes().DropOne(ctx, tc.dropIndexName)
						compatRes, compatErr = compatCollection.Indexes().DropOne(ctx, tc.dropIndexName)
					}

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						t.Logf("Compat error: %v", compatErr)

						AssertMatchesCommandError(t, compatErr, targetErr)
					} else {
						nonEmptyResults = true
						require.NoError(t, compatErr, "compat error; target returned no error")
					}

					require.Equal(t, compatRes, targetRes)

					// List indexes to see they are identical after drop.
					targetCursor, targetErr := targetCollection.Indexes().List(ctx)
					compatCursor, compatErr := compatCollection.Indexes().List(ctx)

					if targetCursor != nil {
						defer targetCursor.Close(ctx)
					}
					if compatCursor != nil {
						defer compatCursor.Close(ctx)
					}

					require.NoError(t, targetErr)
					require.NoError(t, compatErr)

					targetIndexes := FetchAll(t, ctx, targetCursor)
					compatIndexes := FetchAll(t, ctx, compatCursor)

					require.Equal(t, compatIndexes, targetIndexes)
				})
			}

			switch tc.resultType {
			case NonEmptyResult:
				if tc.failsForFerretDB != "" {
					return
				}

				require.True(t, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case EmptyResult:
				require.False(t, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

func TestCreateIndexesCompatUnique(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // for readability
		models           []mongo.IndexModel // required, index to create
		insertDoc        bson.D             // required, document to insert for uniqueness check
		new              bool               // optional, insert new document before check uniqueness
		failsForFerretDB string
	}{
		"IDIndex": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"_id", 1}},
					Options: options.Index().SetUnique(true),
				},
			},
			insertDoc:        bson.D{{"_id", "int322"}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/296",
		},
		"ExistingFieldIndex": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"v", 1}},
					Options: options.Index().SetUnique(true),
				},
			},
			insertDoc: bson.D{{"v", "value"}},
			new:       true,
			// This test case passes only because of our hack for
			// TODO https://github.com/microsoft/documentdb/issues/25
			// failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/296",
		},
		"NotExistingFieldIndex": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"not-existing-field", 1}},
					Options: options.Index().SetUnique(true),
				},
			},
			insertDoc:        bson.D{{"not-existing-field", "value"}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/296",
		},
		"NotUniqueIndex": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"v", 1}},
					Options: options.Index().SetUnique(false),
				},
			},
			insertDoc: bson.D{{"v", "value"}},
		},
		"CompoundIndex": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"v", 1}, {"foo", 1}},
					Options: options.Index().SetUnique(true),
				},
			},
			insertDoc: bson.D{{"v", "baz"}, {"foo", "bar"}},
		},
		"ExistingInsertDuplicate": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"v", 1}},
					Options: options.Index().SetUnique(true),
				},
			},
			insertDoc: bson.D{{"v", int32(42)}},
			// This test case passes only because of our hack for
			// TODO https://github.com/microsoft/documentdb/issues/25
			// failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/296",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt
			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			res := setup.SetupCompatWithOpts(tt,
				&setup.SetupCompatOpts{
					Providers: []shareddata.Provider{shareddata.Int32s},
				})

			ctx, targetCollections, compatCollections := res.Ctx, res.TargetCollections, res.CompatCollections

			targetCollection := targetCollections[0]
			compatCollection := compatCollections[0]

			targetRes, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.models)
			compatRes, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.models)

			if targetErr != nil {
				t.Logf("Target error: %v", targetErr)
				t.Logf("Compat error: %v", compatErr)

				// error messages are intentionally not compared
				AssertMatchesCommandError(t, compatErr, targetErr)

				return
			}
			require.NoError(t, compatErr, "compat error; target returned no error")

			assert.Equal(t, compatRes, targetRes)

			// List specifications to check they are identical after creation.
			targetSpec, targetErr := targetCollection.Indexes().ListSpecifications(ctx)
			compatSpec, compatErr := compatCollection.Indexes().ListSpecifications(ctx)

			require.NoError(t, compatErr)
			require.NoError(t, targetErr)

			assert.Equal(t, compatSpec, targetSpec)

			if tc.new {
				_, targetErr = targetCollection.InsertOne(ctx, tc.insertDoc)
				_, compatErr = compatCollection.InsertOne(ctx, tc.insertDoc)

				require.NoError(t, compatErr)
				require.NoError(t, targetErr)
			}

			_, targetErr = targetCollection.InsertOne(ctx, tc.insertDoc)
			_, compatErr = compatCollection.InsertOne(ctx, tc.insertDoc)

			if targetErr != nil {
				t.Logf("Target error: %v", targetErr)
				t.Logf("Compat error: %v", compatErr)

				// error messages are intentionally not compared
				AssertMatchesWriteError(t, compatErr, targetErr)

				return
			}

			require.NoError(t, compatErr, "compat error; target returned no error")
		})
	}
}

func TestCreateIndexesCompatDuplicates(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // for readability
		models     []mongo.IndexModel       // normal indexes to create
		duplicates []mongo.IndexModel       // duplicates to attempt to create
		resultType CompatTestCaseResultType // defaults to NonEmptyResult
	}{
		"DuplicateByName": {
			models: []mongo.IndexModel{
				{Options: &options.IndexOptions{Name: pointer.To("index_foo")}, Keys: bson.D{{"foo", 1}}},
			},
			duplicates: []mongo.IndexModel{
				{Options: &options.IndexOptions{Name: pointer.To("index_foo")}, Keys: bson.D{{"bar", 1}}},
			},
			resultType: EmptyResult,
		},
		"DuplicateByKey": {
			models: []mongo.IndexModel{
				{Options: &options.IndexOptions{Name: pointer.To("index_foo")}, Keys: bson.D{{"foo", 1}}},
			},
			duplicates: []mongo.IndexModel{
				{Options: &options.IndexOptions{Name: pointer.To("index_not_foo")}, Keys: bson.D{{"foo", 1}}},
			},
			resultType: EmptyResult,
		},
		"DuplicateByPrimaryKey": {
			models: []mongo.IndexModel{
				{Options: new(options.IndexOptions), Keys: bson.D{{"_id", 1}}},
			},
			duplicates: []mongo.IndexModel{
				{Options: &options.IndexOptions{Name: pointer.To("index_foo")}, Keys: bson.D{{"_id", 1}}},
			},
			resultType: EmptyResult,
		},
		"FullyIdentical": {
			models: []mongo.IndexModel{
				{Options: &options.IndexOptions{Name: pointer.To("index_foo")}, Keys: bson.D{{"foo", 1}}},
			},
			duplicates: []mongo.IndexModel{
				{Options: &options.IndexOptions{Name: pointer.To("index_foo")}, Keys: bson.D{{"foo", 1}}},
			},
		},
		"DuplicateAndInvalid": {
			models: []mongo.IndexModel{
				{Options: &options.IndexOptions{Name: pointer.To("index_foo")}, Keys: bson.D{{"foo", 1}}},
			},
			duplicates: []mongo.IndexModel{
				{Options: &options.IndexOptions{Name: pointer.To("index_foo")}, Keys: bson.D{{"foo", 1}}}, // duplicate
				{Options: &options.IndexOptions{Name: pointer.To("index_bar")}, Keys: bson.D{{"foo", 1}}}, // invalid
				{Options: &options.IndexOptions{Name: pointer.To("index_baz")}, Keys: bson.D{{"baz", 1}}}, // valid
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Helper()
			t.Parallel()

			res := setup.SetupCompatWithOpts(t,
				&setup.SetupCompatOpts{
					Providers: []shareddata.Provider{shareddata.Int32s},
				})

			ctx, targetCollections, compatCollections := res.Ctx, res.TargetCollections, res.CompatCollections

			targetCollection := targetCollections[0]
			compatCollection := compatCollections[0]

			targetRes, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.models)
			compatRes, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.models)

			require.NoError(t, compatErr)
			require.NoError(t, targetErr)
			require.Equal(t, compatRes, targetRes)

			targetRes, targetErr = targetCollection.Indexes().CreateMany(ctx, tc.duplicates)
			compatRes, compatErr = compatCollection.Indexes().CreateMany(ctx, tc.duplicates)

			if targetErr != nil {
				t.Logf("Target error: %v", targetErr)
				t.Logf("Compat error: %v", compatErr)

				// error messages are intentionally not compared
				AssertMatchesCommandError(t, compatErr, targetErr)

				return
			}
			require.NoError(t, compatErr, "compat error; target returned no error")

			assert.Equal(t, compatRes, targetRes)
		})
	}
}
